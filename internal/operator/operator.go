package operator

import (
	"flag"
	"fmt"
	"net/http"
	"net/http/pprof"
	"os"
	"strings"
	"time"

	"github.com/mitchellh/mapstructure"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/cluster"
	"sigs.k8s.io/controller-runtime/pkg/healthz"

	stasv1alpha1 "github.com/statnett/image-scanner-operator/api/stas/v1alpha1"
	"github.com/statnett/image-scanner-operator/internal/config"
	"github.com/statnett/image-scanner-operator/internal/controller/stas"
	"github.com/statnett/image-scanner-operator/internal/metrics"
	"github.com/statnett/image-scanner-operator/internal/pod"
	"github.com/statnett/image-scanner-operator/internal/resources"
)

var (
	scheme = runtime.NewScheme()
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	//+kubebuilder:scaffold:scheme
	utilruntime.Must(stasv1alpha1.AddToScheme(scheme))
}

type Operator struct{}

func (o Operator) BindFlags(cfg *config.Config, fs *flag.FlagSet) error {
	fs.String("metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	fs.String("health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	fs.Bool("leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	fs.Bool("enable-profiling", false, "Enable profiling (pprof); available on metrics endpoint.")
	fs.String("namespaces", "", "comma-separated list of namespaces to watch")
	fs.String("cis-metrics-labels", "", "comma-separated list of labels in CIS resources to create metrics labels for")
	fs.Duration("scan-interval", 12*time.Hour, "The minimum time between fetch scan reports from image scanner")
	fs.String("scan-job-namespace", "", "The namespace to schedule scan jobs.")
	fs.String("scan-job-service-account", "default", "The service account used to run scan jobs.")
	fs.String("scan-workload-resources", "", "comma-separated list of workload resources to scan")
	fs.String("scan-namespace-exclude-regexp", "^(kube-|openshift-).*", "regexp for namespace to exclude from scanning")
	fs.String("scan-namespace-include-regexp", "", "regexp for namespace to include for scanning")
	fs.String("trivy-image", "", "The image used for obtaining the trivy binary.")
	fs.Bool("help", false, "print out usage and a summary of options")

	pfs := &pflag.FlagSet{}
	pfs.AddGoFlagSet(fs)

	if err := viper.BindPFlags(pfs); err != nil {
		return fmt.Errorf("unable to bind command line flags: %w", err)
	}

	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))

	return nil
}

func (o Operator) UnmarshalConfig(cfg *config.Config) error {
	helpRequested := viper.GetBool("help")
	if helpRequested {
		pflag.Usage()
		os.Exit(0)
	}

	hook := mapstructure.ComposeDecodeHookFunc(
		mapstructure.StringToTimeDurationHookFunc(),
		mapstructure.StringToSliceHookFunc(","),
		stringToRegexpHookFunc(),
	)
	if err := viper.Unmarshal(cfg, viper.DecodeHook(hook)); err != nil {
		return fmt.Errorf("unable to decode config into struct: %w", err)
	}

	return nil
}

func (o Operator) ValidateConfig(cfg config.Config) error {
	if cfg.ScanJobNamespace == "" {
		return fmt.Errorf("required flag (%q) or env (%q) not set", "scan-job-namespace", "SCAN_JOB_NAMESPACE")
	}

	return nil
}

func (o Operator) Start(cfg config.Config) error {
	metricsAddr := viper.GetString("metrics-bind-address")
	probeAddr := viper.GetString("health-probe-bind-address")
	enableLeaderElection := viper.GetBool("leader-elect")
	options := ctrl.Options{
		NewClient:              cluster.ClientBuilderWithOptions(cluster.ClientOptions{CacheUnstructured: true}),
		Scheme:                 scheme,
		MetricsBindAddress:     metricsAddr,
		Port:                   9443,
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "398aa7bc.statnett.no",
	}

	if len(cfg.ScanNamespaces) > 0 {
		options.NewCache = cache.MultiNamespacedCacheBuilder(cfg.ScanNamespaces)
	}

	kubeConfig := ctrl.GetConfigOrDie()

	mgr, err := ctrl.NewManager(kubeConfig, options)
	if err != nil {
		return fmt.Errorf("unable to start manager: %w", err)
	}

	if err = (&stas.Indexer{}).SetupWithManager(mgr); err != nil {
		return fmt.Errorf("unable to setup indexer: %w", err)
	}

	mapper := &resources.ResourceKindMapper{RestMapper: mgr.GetRESTMapper()}

	kinds, err := mapper.NamespacedKindsForResources(cfg.ScanWorkloadResources...)
	if err != nil {
		return fmt.Errorf("unable to map resources to kinds: %w", err)
	}

	if err = (&stas.PodReconciler{
		Client:        mgr.GetClient(),
		Scheme:        mgr.GetScheme(),
		Config:        cfg,
		WorkloadKinds: kinds,
	}).SetupWithManager(mgr); err != nil {
		return fmt.Errorf("unable to create %s controller: %w", "Pod", err)
	}

	kubeClientset, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		return fmt.Errorf("unable to create Kube ClientSet: %w", err)
	}

	if err = (&stas.ScanJobReconciler{
		Client:     mgr.GetClient(),
		Scheme:     mgr.GetScheme(),
		Config:     cfg,
		LogsReader: pod.NewLogsReader(kubeClientset),
	}).SetupWithManager(mgr); err != nil {
		return fmt.Errorf("unable to create %s controller: %w", "Job", err)
	}

	if err = (&stas.ContainerImageScanReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
		Config: cfg,
	}).SetupWithManager(mgr); err != nil {
		return fmt.Errorf("unable to create %s controller: %w", "ContainerImageScan", err)
	}

	//+kubebuilder:scaffold:builder

	enableProfiling := viper.GetBool("enable-profiling")
	if enableProfiling {
		err = mgr.AddMetricsExtraHandler("/debug/pprof/", http.HandlerFunc(pprof.Index))
		if err != nil {
			return fmt.Errorf("unable to attach pprof to webserver: %w", err)
		}
	}

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		return fmt.Errorf("unable to set up health check: %w", err)
	}

	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		return fmt.Errorf("unable to set up ready check: %w", err)
	}

	if err = (&metrics.ImageMetricsCollector{
		Client: mgr.GetClient(),
		Config: cfg,
	}).SetupWithManager(mgr); err != nil {
		return fmt.Errorf("unable to set up image metrics collector: %w", err)
	}

	ctrl.Log.Info("starting manager")

	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		return fmt.Errorf("problem running manager: %w", err)
	}

	return nil
}
