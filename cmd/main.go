package main

import (
	"flag"
	"net/http"
	"net/http/pprof"
	"os"
	"strings"
	"time"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/statnett/controller-runtime-viper/pkg/zap"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/cluster"
	"sigs.k8s.io/controller-runtime/pkg/healthz"

	stasv1alpha1 "github.com/statnett/image-scanner-operator/api/stas/v1alpha1"
	"github.com/statnett/image-scanner-operator/internal/controller/stas"
	"github.com/statnett/image-scanner-operator/internal/metrics"
	"github.com/statnett/image-scanner-operator/internal/pod"
	"github.com/statnett/image-scanner-operator/internal/resources"
	"github.com/statnett/image-scanner-operator/pkg/operator"
	//+kubebuilder:scaffold:imports
)

const (
	ErrCreateCtrl = "unable to create controller"
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	//+kubebuilder:scaffold:scheme
	utilruntime.Must(stasv1alpha1.AddToScheme(scheme))
}

func main() {
	flag.String("metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.String("health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.Bool("leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	flag.Bool("enable-profiling", false, "Enable profiling (pprof); available on metrics endpoint.")
	flag.String("namespaces", "", "comma-separated list of namespaces to watch")
	flag.String("cis-metrics-labels", "", "comma-separated list of labels in CIS resources to create metrics labels for")
	flag.Duration("scan-interval", 12*time.Hour, "The minimum time between fetch scan reports from image scanner")
	flag.String("scan-job-namespace", "", "The namespace to schedule scan jobs.")
	flag.String("scan-job-service-account", "default", "The service account used to run scan jobs.")
	flag.String("scan-workload-resources", "", "comma-separated list of workload resources to scan")
	flag.String("trivy-image", "", "The image used for obtaining the trivy binary.")
	flag.String("trivy-server", "", "The server to use in Trivy client/server mode.")
	flag.Bool("help", false, "print out usage and a summary of options")

	opts := zap.Options{
		Development: true,
	}
	opts.BindFlags(flag.CommandLine)
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()

	err := viper.BindPFlags(pflag.CommandLine)
	if err != nil {
		setupLog.Error(err, "unable to bind command line flags")
		os.Exit(1)
	}

	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))

	if viper.GetBool("help") {
		pflag.Usage()
		os.Exit(0)
	}

	cfg := operator.Config{}
	if err := viper.Unmarshal(&cfg); err != nil {
		setupLog.Error(err, "unable to decode config into struct")
		os.Exit(1)
	}

	if cfg.ScanJobNamespace == "" {
		setupLog.V(0).Info("required flag/env not set", "flag", "scan-job-namespace", "env", "SCAN_JOB_NAMESPACE")
		os.Exit(1)
	}

	logger := zap.New(zap.UseFlagOptions(&opts))
	ctrl.SetLogger(logger)
	klog.SetLogger(logger)

	options := ctrl.Options{
		NewClient:              cluster.ClientBuilderWithOptions(cluster.ClientOptions{CacheUnstructured: true}),
		Scheme:                 scheme,
		MetricsBindAddress:     viper.GetString("metrics-bind-address"),
		Port:                   9443,
		HealthProbeBindAddress: viper.GetString("health-probe-bind-address"),
		LeaderElection:         viper.GetBool("leader-elect"),
		LeaderElectionID:       "398aa7bc.statnett.no",
	}

	namespaces := viper.GetStringSlice("namespaces")
	if len(namespaces) > 0 {
		options.NewCache = cache.MultiNamespacedCacheBuilder(namespaces)
	}

	kubeConfig := ctrl.GetConfigOrDie()

	mgr, err := ctrl.NewManager(kubeConfig, options)
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	if err = (&stas.Indexer{}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to setup indexer")
		os.Exit(1)
	}

	mapper := &resources.ResourceKindMapper{RestMapper: mgr.GetRESTMapper()}

	kinds, err := mapper.NamespacedKindsForResources(cfg.ScanWorkloadResources...)
	if err != nil {
		setupLog.Error(err, "unable to map resources to kinds")
		os.Exit(1)
	}

	if err = (&stas.PodReconciler{
		Client:        mgr.GetClient(),
		Scheme:        mgr.GetScheme(),
		Config:        cfg,
		WorkloadKinds: kinds,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, ErrCreateCtrl, "controller", "Pod")
		os.Exit(1)
	}

	kubeClientset, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		setupLog.Error(err, "unable to create Kube ClientSet")
		os.Exit(1)
	}

	if err = (&stas.ScanJobReconciler{
		Client:     mgr.GetClient(),
		Scheme:     mgr.GetScheme(),
		Config:     cfg,
		LogsReader: pod.NewLogsReader(kubeClientset),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, ErrCreateCtrl, "controller", "Job")
		os.Exit(1)
	}

	if err = (&stas.ContainerImageScanReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
		Config: cfg,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, ErrCreateCtrl, "controller", "ContainerImageScan")
		os.Exit(1)
	}

	//+kubebuilder:scaffold:builder

	if viper.GetBool("enable-profiling") {
		err = mgr.AddMetricsExtraHandler("/debug/pprof/", http.HandlerFunc(pprof.Index))
		if err != nil {
			setupLog.Error(err, "unable to attach pprof to webserver")
			os.Exit(1)
		}
	}

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}

	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	cisMetricsLabels := viper.GetStringSlice("cis-metrics-labels")
	if err = (&metrics.ImageMetricsCollector{
		Client: mgr.GetClient(),
		Config: cfg,
	}).SetupWithManager(mgr, cisMetricsLabels...); err != nil {
		setupLog.Error(err, "unable to set up image metrics collector")
		os.Exit(1)
	}

	setupLog.Info("starting manager")

	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
