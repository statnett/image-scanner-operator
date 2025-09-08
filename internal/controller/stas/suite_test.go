package stas

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/format"
	openreportsv1alpha1 "github.com/openreports/reports-api/apis/openreports.io/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/json"
	"k8s.io/apimachinery/pkg/util/uuid"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/envtest/komega"
	"sigs.k8s.io/controller-runtime/pkg/event"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/kustomize/api/krusty"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/filesys"

	stasv1alpha1 "github.com/statnett/image-scanner-operator/api/stas/v1alpha1"
	"github.com/statnett/image-scanner-operator/internal/config"
	"github.com/statnett/image-scanner-operator/internal/config/feature"
	"github.com/statnett/image-scanner-operator/internal/pod"
)

const scanJobNamespace = "image-scanner"

var (
	k8sClient        client.Client
	k8sScheme        *runtime.Scheme
	k8sEventRecorder record.EventRecorder
	testEnv          *envtest.Environment
	ctx              context.Context
	cancel           context.CancelFunc
	logsReader       = new(pod.MockLogsReader)
)

func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecs(t, "Controller Suite")
}

var _ = BeforeSuite(func() {
	// Disable Gomega max length (default is 4000, and that is not enough)
	format.MaxLength = 0
	// Increase Gomega max depth (default is 10, and that is not enough for Deployment)
	format.MaxDepth = 20

	SetDefaultEventuallyTimeout(2 * time.Second)
	SetDefaultConsistentlyDuration(2 * time.Second)

	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))

	ctx, cancel = context.WithCancel(context.TODO())

	Expect(config.DefaultMutableFeatureGate.OverrideDefault(feature.PolicyReport, true)).To(Succeed())

	openReportsCRDs := loadOpenReportsCRDs()

	By("bootstrapping test environment")
	testEnv = &envtest.Environment{
		CRDDirectoryPaths: []string{
			filepath.Join("..", "..", "..", "config", "crd", "bases"),
		},
		CRDs:                  openReportsCRDs,
		ErrorIfCRDPathMissing: true,
	}

	// cfg is defined in this file globally.
	cfg, err := testEnv.Start()
	Expect(err).NotTo(HaveOccurred())
	Expect(cfg).NotTo(BeNil())

	Expect(appsv1.AddToScheme(scheme.Scheme)).To(Succeed())
	Expect(stasv1alpha1.AddToScheme(scheme.Scheme)).To(Succeed())
	Expect(batchv1.AddToScheme(scheme.Scheme)).To(Succeed())
	Expect(openreportsv1alpha1.Install(scheme.Scheme)).To(Succeed())

	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
	Expect(err).NotTo(HaveOccurred())
	Expect(k8sClient).NotTo(BeNil())
	komega.SetClient(k8sClient)

	namespaces := []string{scanJobNamespace, "replica-set", "stateful-set"}
	for _, name := range namespaces {
		namespace := &corev1.Namespace{}
		namespace.Name = name
		Expect(k8sClient.Create(ctx, namespace)).To(Succeed())
	}

	k8sManager, err := ctrl.NewManager(cfg, ctrl.Options{
		Client: client.Options{Cache: &client.CacheOptions{Unstructured: true}},
		Scheme: scheme.Scheme,
	})
	Expect(err).NotTo(HaveOccurred())

	k8sEventRecorder = k8sManager.GetEventRecorderFor("test")
	Expect(k8sEventRecorder).NotTo(BeNil())

	indexer := &Indexer{}
	Expect(indexer.SetupWithManager(k8sManager)).To(Succeed())

	k8sScheme = k8sManager.GetScheme()

	config := config.Config{
		ReuseScanResults:               true,
		ScanJobNamespace:               scanJobNamespace,
		ScanJobServiceAccount:          "image-scanner-job",
		ScanJobTTLSecondsAfterFinished: 60,
		ScanInterval:                   time.Hour,
		TrivyCommand:                   config.RootfsTrivyCommand,
		TrivyImage:                     "aquasecurity/trivy",
	}

	podReconciler := &PodReconciler{
		Client: k8sManager.GetClient(),
		Scheme: k8sScheme,
		Config: config,
		WorkloadKinds: []schema.GroupVersionKind{
			{Group: "apps", Version: "v1", Kind: "ReplicaSet"},
			{Group: "apps", Version: "v1", Kind: "StatefulSet"},
		},
	}
	Expect(podReconciler.SetupWithManager(k8sManager)).To(Succeed())

	cisEventChan := make(chan event.GenericEvent)
	containerImageScanReconciler := &ContainerImageScanReconciler{
		Client:    k8sManager.GetClient(),
		Scheme:    k8sScheme,
		Config:    config,
		EventChan: cisEventChan,
	}
	Expect(containerImageScanReconciler.SetupWithManager(k8sManager)).To(Succeed())
	rescanTrigger := &RescanTrigger{
		Client:        k8sManager.GetClient(),
		Config:        config,
		EventChan:     cisEventChan,
		CheckInterval: time.Second,
	}
	Expect(k8sManager.Add(rescanTrigger)).To(Succeed())

	scanJobReconciler := &ScanJobReconciler{
		Client:     k8sManager.GetClient(),
		Scheme:     k8sScheme,
		Config:     config,
		LogsReader: logsReader,
	}
	Expect(scanJobReconciler.SetupWithManager(k8sManager)).To(Succeed())

	go func() {
		defer GinkgoRecover()
		var ctrlCtx context.Context
		ctrlCtx, cancel = context.WithCancel(ctrl.SetupSignalHandler())
		Expect(k8sManager.Start(ctrlCtx)).To(Succeed())
	}()
})

func loadOpenReportsCRDs() []*apiextensionsv1.CustomResourceDefinition {
	kOpts := krusty.MakeDefaultOptions()
	kOpts.PluginConfig = types.EnabledPluginConfig(types.BploUndefined)
	kOpts.PluginConfig.HelmConfig.Command = "helm"
	k := krusty.MakeKustomizer(kOpts)
	m, err := k.Run(filesys.FileSystemOrOnDisk{}, filepath.Join("..", "..", "..", "config", "openreports"))
	Expect(err).To(Succeed())

	resources := m.Resources()

	crds := make([]*apiextensionsv1.CustomResourceDefinition, len(resources))
	for i := range resources {
		bytes, err := resources[i].MarshalJSON()
		Expect(err).To(Succeed())

		crd := &apiextensionsv1.CustomResourceDefinition{}
		err = json.Unmarshal(bytes, crd)
		Expect(err).To(Succeed())

		crds[i] = crd
	}

	return crds
}

var _ = AfterSuite(func() {
	cancel()

	By("tearing down the test environment")
	err := testEnv.Stop()
	Expect(err).NotTo(HaveOccurred())
})

func createPod(ctx context.Context, owner client.Object, s *runtime.Scheme) *corev1.Pod {
	p := newPod(owner, s)
	podCopy := p.DeepCopy()
	Expect(k8sClient.Create(ctx, podCopy)).To(Succeed())
	Expect(k8sClient.Status().Update(ctx, p)).To(Succeed())

	return p
}

func newPod(owner client.Object, s *runtime.Scheme) *corev1.Pod {
	p := &corev1.Pod{}
	p.Namespace = owner.GetNamespace()
	p.Name = owner.GetName() + string(uuid.NewUUID())
	p.Labels = owner.GetLabels()
	err := controllerutil.SetControllerReference(owner, p, s)
	Expect(err).To(Succeed())

	p.Spec.Containers = []corev1.Container{
		{
			Name:  "foo",
			Image: "my.registry/repository/app:f54a333e",
		},
	}
	p.Status.ContainerStatuses = []corev1.ContainerStatus{
		{
			Name:    "foo",
			Image:   "my.registry/repository/app:f54a333e",
			ImageID: "my.registry/repository/app@sha256:4b59f7dacd37c688968756d176139715df69d89eb0be1802e059316f9d58d9ef",
		},
	}
	setPodReady(p)

	return p
}

func setPodReady(p *corev1.Pod) {
	// Simulate Pod kstatus Current
	p.Status.Phase = corev1.PodRunning
	p.Status.Conditions = []corev1.PodCondition{{
		Type:   corev1.PodReady,
		Status: corev1.ConditionTrue,
	}}
}
