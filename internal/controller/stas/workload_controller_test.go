package stas

import (
	"context"
	"sort"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/envtest/komega"

	stasv1alpha1 "github.com/statnett/image-scanner-operator/api/stas/v1alpha1"
)

const DefaultNamespaceName = "default"

type TestWorkloadFactory func(namespacedName types.NamespacedName, labels map[string]string) client.Object

var _ = Describe("Workload controller", func() {
	const (
		timeout  = 5 * time.Second
		interval = 100 * time.Millisecond
	)

	var (
		ctx context.Context
	)

	BeforeEach(func() {
		ctx = context.Background()
	})

	DescribeTable("should create ContainerImageScan with pod image from matching pod",
		func(namespace string, workloadFactory TestWorkloadFactory) {
			namespacedName := types.NamespacedName{
				Namespace: namespace,
				Name:      "matching-pods",
			}

			labels := map[string]string{"name": namespacedName.Name}

			workload := workloadFactory(namespacedName, labels)
			Expect(k8sClient.Create(ctx, workload)).To(Succeed())

			createPod(ctx, workload, k8sClient.Scheme())
			expectedImage := stasv1alpha1.Image{
				Name:   "my.registry/repository/app",
				Digest: "sha256:4b59f7dacd37c688968756d176139715df69d89eb0be1802e059316f9d58d9ef",
			}

			imageScans := &stasv1alpha1.ContainerImageScanList{}
			listOps := []client.ListOption{
				client.InNamespace(namespacedName.Namespace),
				client.MatchingLabels(labels),
			}
			Eventually(komega.ObjectList(imageScans, listOps...), timeout, interval).Should(HaveField("Items", HaveLen(1)))
			Expect(imageScans.Items[0].Spec.Image).To(Equal(expectedImage))
			Expect(imageScans.Items[0].Spec.Tag).To(Equal("f54a333e"))
		},
		Entry("ReplicaSet", "replica-set", newReplicaSet),
		Entry("StatefulSet", "stateful-set", newStatefulSet),
	)

	It("should pick up ignore-unfixed annotation from workload", func() {
		pod := &corev1.Pod{}
		pod.Namespace = DefaultNamespaceName
		pod.Name = "ignore-unfixed"
		pod.Annotations = map[string]string{"image-scanner.statnett.no/ignore-unfixed": "true"}
		pod.Labels = map[string]string{"app": "oauth2-proxy"}
		pod.Spec.Containers = []corev1.Container{
			{
				Name:  "oauth-proxy",
				Image: "quay.io/oauth2-proxy/oauth2-proxy",
			},
		}
		pod.Status.ContainerStatuses = []corev1.ContainerStatus{
			{
				Image:   "quay.io/oauth2-proxy/oauth2-proxy:latest",
				ImageID: "quay.io/oauth2-proxy/oauth2-proxy@sha256:10615e4f03bddba4cd49823420d9f50a403776d1b58991caa6d123e3527ff79f",
			},
		}
		setPodReady(pod)
		Expect(k8sClient.Create(ctx, pod.DeepCopy())).To(Succeed())
		Expect(k8sClient.Status().Update(ctx, pod)).To(Succeed())

		imageScans := &stasv1alpha1.ContainerImageScanList{}
		listOps := []client.ListOption{
			client.InNamespace(pod.Namespace),
			client.MatchingLabels(pod.Labels),
		}
		Eventually(komega.ObjectList(imageScans, listOps...), timeout, interval).Should(HaveField("Items", HaveLen(1)))
		Expect(imageScans.Items[0].Spec.IgnoreUnfixed).To(Equal(ptr.To(true)))
	})

	It("should add all Pods from same workload with same image as CIS owners", func() {
		newReplicaset := func(name string) *appsv1.ReplicaSet {
			rs := &appsv1.ReplicaSet{}
			rs.Namespace = DefaultNamespaceName
			rs.Name = name
			rs.Spec.Template.Labels = map[string]string{"app": name, "test": "controller"}
			rs.Spec.Template.Spec.Containers = []corev1.Container{
				{
					Name:  "oauth-proxy-1",
					Image: "quay.io/oauth2-proxy/oauth2-proxy",
				},
				{
					Name:  "oauth-proxy-2",
					Image: "quay.io/oauth2-proxy/oauth2-proxy",
				},
			}
			rs.Spec.Selector = &metav1.LabelSelector{
				MatchLabels: rs.Spec.Template.Labels,
			}
			return rs
		}

		newPod := func(rs *appsv1.ReplicaSet, name string, sha string, containerSuffix string) *corev1.Pod {
			pod := &corev1.Pod{}
			pod.Namespace = rs.Namespace
			pod.Name = name
			pod.Labels = rs.Spec.Selector.MatchLabels
			// containerSuffix allows simulating changing container names in
			// ReplicaSet and having Pods with same owner and differende
			// container names.
			for _, c := range rs.Spec.Template.Spec.Containers {
				pod.Spec.Containers = append(pod.Spec.Containers,
					corev1.Container{
						Name:  c.Name + containerSuffix,
						Image: c.Image,
					},
				)
				pod.Status.ContainerStatuses = append(pod.Status.ContainerStatuses,
					corev1.ContainerStatus{
						Name:    c.Name + containerSuffix,
						Image:   c.Image + ":latest",
						ImageID: c.Image + "@" + sha,
					},
				)
			}
			setPodReady(pod)

			Expect(controllerutil.SetControllerReference(rs, pod, testEnv.Scheme)).To(Succeed())
			return pod
		}

		rs1 := newReplicaset("controller-1")
		rs2 := newReplicaset("controller-2")

		Expect(k8sClient.Create(ctx, rs1)).To(Succeed())
		Expect(k8sClient.Create(ctx, rs2)).To(Succeed())

		const FirstSHA = "sha256:10615e4f03bddba4cd49823420d9f50a403776d1b58991caa6d123e3527ff79f"
		const SecondSHA = "sha256:45dddaa9b519329a688366e2b6119214a42cac569529ccacb0989c43355f0255"
		pod1 := newPod(rs1, "controlled-1", FirstSHA, "")
		pod2 := newPod(rs1, "controlled-2", FirstSHA, "")
		// This pod simulates the ReplicaSet previously having different
		// container names. As the Replica set's container names differ in
		// differnt pods, multiple ContainerImageScans should be created.
		pod3 := newPod(rs1, "controlled-3", FirstSHA, "-old-rs")
		pod4 := newPod(rs1, "controlled-4", SecondSHA, "")
		pod5 := newPod(rs2, "controlled-5", SecondSHA, "")

		Expect(k8sClient.Create(ctx, pod1.DeepCopy())).To(Succeed())
		Expect(k8sClient.Create(ctx, pod2.DeepCopy())).To(Succeed())
		Expect(k8sClient.Create(ctx, pod3.DeepCopy())).To(Succeed())
		Expect(k8sClient.Create(ctx, pod4.DeepCopy())).To(Succeed())
		Expect(k8sClient.Create(ctx, pod5.DeepCopy())).To(Succeed())
		Expect(k8sClient.Status().Update(ctx, pod1)).To(Succeed())
		Expect(k8sClient.Status().Update(ctx, pod2)).To(Succeed())
		Expect(k8sClient.Status().Update(ctx, pod3)).To(Succeed())
		Expect(k8sClient.Status().Update(ctx, pod4)).To(Succeed())
		Expect(k8sClient.Status().Update(ctx, pod5)).To(Succeed())

		ownerRefTransform := func(imageScans *stasv1alpha1.ContainerImageScanList) [][]types.UID {
			ownerRefs := make([][]types.UID, 0, len(imageScans.Items))
			sort.Slice(imageScans.Items, func(i, j int) bool {
				return imageScans.Items[i].Name < imageScans.Items[j].Name
			})
			for i := range imageScans.Items {
				sort.Slice(imageScans.Items[i].OwnerReferences, func(j, k int) bool {
					return imageScans.Items[i].OwnerReferences[j].Name < imageScans.Items[i].OwnerReferences[k].Name
				})
				ors := make([]types.UID, 0, len(imageScans.Items[i].OwnerReferences))
				for _, or := range imageScans.Items[i].OwnerReferences {
					ors = append(ors, or.UID)
				}
				ownerRefs = append(ownerRefs, ors)
			}
			return ownerRefs
		}

		Eventually(
			komega.ObjectList(&stasv1alpha1.ContainerImageScanList{},
				client.InNamespace(rs1.Namespace),
				client.MatchingLabels(map[string]string{"test": "controller"}),
			), timeout, interval,
		).Should(WithTransform(ownerRefTransform, Equal([][]types.UID{
			{pod4.UID},
			{pod1.UID, pod2.UID}, // Not pod3 due to it having different container names
			{pod3.UID},
			{pod4.UID},
			{pod1.UID, pod2.UID},
			{pod3.UID},
			{pod5.UID},
			{pod5.UID},
		})))
	})

	It("should delete obsolete ContainerImageScan", func() {
		pod := &corev1.Pod{}
		pod.Namespace = DefaultNamespaceName
		pod.Name = "crashing-pod"
		pod.Labels = map[string]string{"app": "crashing-pod"}
		pod.Spec.Containers = []corev1.Container{
			{
				Name:  "app",
				Image: "foo-app",
			},
		}
		pod.Status.ContainerStatuses = []corev1.ContainerStatus{
			{
				Image:   "dummy.registry.mycorp.com/foo-app:latest",
				ImageID: "dummy.registry.mycorp.com/foo-app@sha256:45dddaa9b519329a688366e2b6119214a42cac569529ccacb0989c43355f0255",
			},
		}
		setPodReady(pod)
		Expect(k8sClient.Create(ctx, pod.DeepCopy())).To(Succeed())
		Expect(k8sClient.Status().Update(ctx, pod)).To(Succeed())

		imageScans := &stasv1alpha1.ContainerImageScanList{}
		listOps := []client.ListOption{
			client.InNamespace(pod.Namespace),
			client.MatchingLabels(pod.Labels),
		}
		Eventually(komega.ObjectList(imageScans, listOps...), timeout, interval).Should(HaveField("Items", HaveLen(1)))
		cis := &imageScans.Items[0]

		// Update ImageID with new digest
		pod.Status.ContainerStatuses[0].ImageID = "dummy.registry.mycorp.com/foo-app@sha256:8dda7152241873a583062c925694f1a2f5cdf1bc1e40df57ef598e2520ef09f6"
		Expect(k8sClient.Status().Update(ctx, pod)).To(Succeed())

		// Assert obsolete CIS deleted
		Eventually(komega.Get(cis), timeout, interval).Should(WithTransform(errors.ReasonForError, Equal(metav1.StatusReasonNotFound)))

		// Assert new CIS present
		Eventually(komega.ObjectList(imageScans, listOps...), timeout, interval).Should(HaveField("Items", HaveLen(1)))
		cis2 := &imageScans.Items[0]
		Expect(cis2.Name).To(Not(Equal(cis.Name)))
	})
})

var _ = Describe("Naming ContainerImageScan", func() {
	var img stasv1alpha1.Image
	var ctrl client.Object
	var containerName string

	BeforeEach(func() {
		ctrl = &metav1.PartialObjectMetadata{
			TypeMeta: metav1.TypeMeta{
				Kind: "Application",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: "workload",
			},
		}
		containerName = "app"
		img = stasv1alpha1.Image{Name: "img-name", Digest: "img-digest"}
		Expect(imageScanName(ctrl, containerName, img)).To(Equal("application-workload-app-88c48"))
	})

	It("should contain controller name", func() {
		ctrl.SetName("other-workload")
		Expect(imageScanName(ctrl, containerName, img)).To(Equal("application-other-workload-app-88c48"))
	})

	It("should be a function of image name", func() {
		img.Name = "other-img"
		Expect(imageScanName(ctrl, containerName, img)).To(Equal("application-workload-app-91ac0"))
	})

	It("should be a function of image digest", func() {
		img.Digest = "other-digest"
		Expect(imageScanName(ctrl, containerName, img)).To(Equal("application-workload-app-faf6e"))
	})

	It("should contain container name", func() {
		containerName = "foo"
		Expect(imageScanName(ctrl, containerName, img)).To(Equal("application-workload-foo-88c48"))
	})

	It("should contain controller kind", func() {
		ctrl = &metav1.PartialObjectMetadata{
			TypeMeta: metav1.TypeMeta{
				Kind: "Deployment",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: ctrl.GetName(),
			},
		}
		Expect(imageScanName(ctrl, containerName, img)).To(Equal("deployment-workload-app-88c48"))
	})

	It("should shorten controller name part", func() {
		longName := "kexovomawokadivasuketalewayonepiseziqaqitotasenenegekayerucugasojalunenuherejepetemutotacoyeyotuxutesereratowitanedeviwetelecifokoxoviwonejiraroxasohohacamariserilasecehoreratisetabamocanobotuwocosorehohonatonatehohenatixacinotanicinocerurazawilemupisir"
		Expect(longName).To(HaveLen(KubernetesNameMaxLength))

		ctrl.SetName(longName)
		cisName, err := imageScanName(ctrl, containerName, img)
		Expect(err).NotTo(HaveOccurred())
		Expect(cisName).To(HaveLen(KubernetesNameMaxLength))
		// Assert contains image short sha part
		Expect(cisName).To(ContainSubstring("-88c48"))
	})
})

func newReplicaSet(namespacedName types.NamespacedName, labels map[string]string) client.Object {
	rs := &appsv1.ReplicaSet{}
	rs.Namespace = namespacedName.Namespace
	rs.Name = namespacedName.Name
	rs.Labels = labels
	rs.Spec.Selector = &metav1.LabelSelector{
		MatchLabels: labels,
	}
	rs.Spec.Template.Labels = labels
	rs.Spec.Template.Spec.Containers = []corev1.Container{
		{
			Name:  "foo",
			Image: "foo-image",
		},
	}

	return rs
}

func newStatefulSet(namespacedName types.NamespacedName, labels map[string]string) client.Object {
	ss := &appsv1.StatefulSet{}
	ss.Namespace = namespacedName.Namespace
	ss.Name = namespacedName.Name
	ss.Labels = labels
	ss.Spec.Selector = &metav1.LabelSelector{
		MatchLabels: labels,
	}
	ss.Spec.Template.Labels = labels
	ss.Spec.Template.Spec.Containers = []corev1.Container{
		{
			Name:  "foo",
			Image: "foo-image",
		},
	}

	return ss
}
