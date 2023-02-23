package stas

import (
	"context"
	"fmt"
	"path"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/envtest/komega"

	stasv1alpha1 "github.com/statnett/image-scanner-operator/api/stas/v1alpha1"
	"github.com/statnett/image-scanner-operator/internal/yaml"
)

var _ = Describe("Scan Job BackOff Pod controller", func() {
	BeforeEach(func() {
		ctx = context.Background()
	})

	Context("When scan job pod is backing off", func() {
		It("should delete job and update CIS status", func() {
			// Create CIS
			cis := &stasv1alpha1.ContainerImageScan{}
			Expect(yaml.FromFile(path.Join("testdata", "scan-job-backoff-pod", "cis.yaml"), cis)).To(Succeed())
			Expect(k8sClient.Create(ctx, cis)).To(Succeed())

			// Wait for CIS to get reconciled
			Eventually(komega.Object(cis)).Should(HaveField("Status.ObservedGeneration", Not(BeZero())))
			// Sanity check for conditions set
			Expect(cis.Status.Conditions).To(Not(BeEmpty()))

			// Simulate scan job pod backoff
			scanJob := getContainerImageScanJob(cis)
			pod := createScanJobPod(scanJob)
			Expect(komega.UpdateStatus(pod, func() {
				pod.Status.ContainerStatuses = []corev1.ContainerStatus{{
					State: corev1.ContainerState{
						Waiting: &corev1.ContainerStateWaiting{
							Reason:  "ErrImagePull",
							Message: "Image not found",
						},
					},
				}}
			})()).To(Succeed())
			k8sEventRecorder.Event(pod, corev1.EventTypeNormal, "BackOff", fmt.Sprintf("Back-off pulling image %q", cis.Spec.Digest))

			// Wait for back-off scan job to get reconciled
			Eventually(komega.Object(cis)).Should(HaveField("Status.LastScanTime", Not(BeZero())))
			Expect(cis.Status.LastSuccessfulScanTime).To(BeZero())
			Expect(cis.Status.LastScanJobUID).To(Equal(scanJob.UID))
			// Check conditions
			Expect(cis.Status.Conditions).To(HaveLen(1))
			condition := cis.Status.Conditions[0]
			Expect(condition.Type).To(Equal("Stalled"))
			Expect(condition.Status).To(Equal(metav1.ConditionTrue))
			Expect(condition.Reason).To(Equal("Error"))
			Expect(condition.Message).To(Equal("Image not found"))

			// Check that scan job is "deleted"
			Eventually(komega.Object(scanJob)).Should(HaveField("ObjectMeta.DeletionTimestamp", Not(BeZero())))
		})
	})
})
