package controllers

import (
	"context"
	"path/filepath"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/format"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/uuid"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/cluster"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/envtest/komega"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	stasv1alpha1 "github.com/statnett/image-scanner-operator/api/v1alpha1"
	"github.com/statnett/image-scanner-operator/internal/pod"
	"github.com/statnett/image-scanner-operator/pkg/operator"
	//+kubebuilder:scaffold:imports
)

const scanJobNamespace = "image-scanner-jobs"

var (
	cfg        *rest.Config
	k8sClient  client.Client
	k8sScheme  *runtime.Scheme
	testEnv    *envtest.Environment
	ctx        context.Context
	cancel     context.CancelFunc
	logsReader = new(pod.MockLogsReader)
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

	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))

	ctx, cancel = context.WithCancel(context.TODO())

	By("bootstrapping test environment")
	testEnv = &envtest.Environment{
		CRDDirectoryPaths:     []string{filepath.Join("..", "config", "crd", "bases")},
		ErrorIfCRDPathMissing: true,
	}

	var err error
	// cfg is defined in this file globally.
	cfg, err = testEnv.Start()
	Expect(err).NotTo(HaveOccurred())
	Expect(cfg).NotTo(BeNil())

	Expect(appsv1.AddToScheme(scheme.Scheme)).To(Succeed())
	Expect(stasv1alpha1.AddToScheme(scheme.Scheme)).To(Succeed())
	Expect(batchv1.AddToScheme(scheme.Scheme)).To(Succeed())
	//+kubebuilder:scaffold:scheme

	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
	Expect(err).NotTo(HaveOccurred())
	Expect(k8sClient).NotTo(BeNil())
	komega.SetClient(k8sClient)

	namespaces := []string{scanJobNamespace, "replica-set", "stateful-set"}
	for _, name := range namespaces {
		namespace := &corev1.Namespace{}
		namespace.Name = name
		err := k8sClient.Create(ctx, namespace)
		Expect(err).To(Succeed())
	}

	// FIXME: temporarily disable cache to see if the tests pass in Github
	uncachedObjects := []client.Object{&batchv1.Job{}}
	k8sManager, err := ctrl.NewManager(cfg, ctrl.Options{
		NewClient: cluster.ClientBuilderWithOptions(cluster.ClientOptions{CacheUnstructured: true, UncachedObjects: uncachedObjects}),
		Scheme:    scheme.Scheme,
	})
	Expect(err).NotTo(HaveOccurred())

	indexer := &Indexer{}
	Expect(indexer.SetupWithManager(k8sManager)).To(Succeed())

	k8sScheme = k8sManager.GetScheme()

	config := operator.Config{
		ScanJobNamespace:      scanJobNamespace,
		ScanJobServiceAccount: "image-scanner",
		TrivyImage:            "aquasecurity/trivy",
		TrivyServer:           "http://trivy.image-scanner.svc.cluster.local",
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

	containerImageScanReconciler := &ContainerImageScanReconciler{
		Client: k8sManager.GetClient(),
		Scheme: k8sScheme,
		Config: config,
	}
	Expect(containerImageScanReconciler.SetupWithManager(k8sManager)).To(Succeed())

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
			Image:   "my.registry/repository/app:f54a333e",
			ImageID: "my.registry/repository/app@sha256:4b59f7dacd37c688968756d176139715df69d89eb0be1802e059316f9d58d9ef",
		},
	}

	return p
}
