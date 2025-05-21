package operator

import (
	"flag"
	"fmt"
	"net/http"
	"net/http/pprof"
	"os"
	"strings"
	"time"

	"github.com/go-viper/mapstructure/v2"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	eventsv1 "k8s.io/api/events/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	ctrlconfig "sigs.k8s.io/controller-runtime/pkg/config"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/metrics/server"
	policyv1alpha2 "sigs.k8s.io/wg-policy-prototypes/policy-report/apis/wgpolicyk8s.io/v1alpha2"

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
	utilruntime.Must(stasv1alpha1.AddToScheme(scheme))
	utilruntime.Must(policyv1alpha2.AddToScheme(scheme))
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
	fs.Int("scan-job-ttl-seconds-after-finished", 7200, "The lifetime (in seconds) of a scan job that has finished. Value must be positive to allow scan reports to be harvested by the operator.")
	fs.String("scan-workload-resources", "", "A comma-separated list of workload resources to scan. Format used for resource is \"resource.group\", i.e. \"deployments.apps\".")
	fs.String("scan-namespace-exclude-regexp", "^(kube-|openshift-).*", "regexp for namespace to exclude from scanning")
	fs.String("scan-namespace-include-regexp", "", "regexp for namespace to include for scanning")
	fs.String("skip-scan-pod-waiting-reasons", "", "A comma-separated list of pod reasons that should result in skipping scan. ErrImagePull and ImagePullBackOff are skipped regardless of config.")
	fs.String("trivy-command", string(config.RootfsTrivyCommand), "The trivy command used to scan filesystem in image; can be 'filesystem' or 'rootfs'")
	fs.String("trivy-image", "", "The image used for obtaining the trivy binary.")
	fs.Int("active-scan-job-limit", 8, "The maximum number of active scan jobs. Setting it to 0 will disable the limit.")
	fs.Bool("reuse-scan-results", false, "Reuse latest valid scan result within interval when possible, instead of starting new scan.")
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
		mapstructure.TextUnmarshallerHookFunc(),
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

	if cfg.ScanJobTTLSecondsAfterFinished <= 0 {
		return fmt.Errorf("flag (%q) or env (%q) must be greater than zero", "scan-job-ttl-seconds-after-finished", "SCAN_JOB_TTL_SECONDS_AFTER_FINISHED")
	}

	return nil
}

func (o Operator) Start(cfg config.Config) error {
	metricsAddr := viper.GetString("metrics-bind-address")
	probeAddr := viper.GetString("health-probe-bind-address")
	enableLeaderElection := viper.GetBool("leader-elect")

	metricsOpts := server.Options{BindAddress: metricsAddr}
	if viper.GetBool("enable-profiling") {
		metricsOpts.ExtraHandlers = map[string]http.Handler{"/debug/pprof/": http.HandlerFunc(pprof.Index)}
	}

	options := ctrl.Options{
		Client: client.Options{Cache: &client.CacheOptions{
			Unstructured: true,
			DisableFor:   []client.Object{&eventsv1.Event{}},
		}},
		Controller: ctrlconfig.Controller{
			UsePriorityQueue: ptr.To(true),
		},
		Scheme:                 scheme,
		MapperProvider:         apiutil.NewDynamicRESTMapper,
		Metrics:                metricsOpts,
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "398aa7bc.statnett.no",
	}

	if len(cfg.ScanNamespaces) > 0 {
		options.Cache.DefaultNamespaces = make(map[string]cache.Config, len(cfg.ScanNamespaces))
		for _, n := range cfg.ScanNamespaces {
			options.Cache.DefaultNamespaces[n] = cache.Config{}
		}
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

	rescanEventChan := make(chan event.GenericEvent)

	if err = (&stas.ContainerImageScanReconciler{
		Client:    mgr.GetClient(),
		Scheme:    mgr.GetScheme(),
		Config:    cfg,
		EventChan: rescanEventChan,
	}).SetupWithManager(mgr); err != nil {
		return fmt.Errorf("unable to create %s controller: %w", "ContainerImageScan", err)
	}

	//+kubebuilder:scaffold:builder

	if err := mgr.Add(&stas.RescanTrigger{
		Client:        mgr.GetClient(),
		Config:        cfg,
		EventChan:     rescanEventChan,
		CheckInterval: time.Minute,
	}); err != nil {
		return fmt.Errorf("unable to create rescan trigger: %w", err)
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
