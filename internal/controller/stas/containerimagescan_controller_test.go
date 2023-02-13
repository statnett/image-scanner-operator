package stas

import (
	"context"
	"path"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/envtest/komega"

	stasv1alpha1 "github.com/statnett/image-scanner-operator/api/stas/v1alpha1"
	"github.com/statnett/image-scanner-operator/internal/trivy"
	"github.com/statnett/image-scanner-operator/internal/yaml"
)

var _ = Describe("ContainerImageScan controller", func() {
	BeforeEach(func() {
		ctx = context.Background()
	})

	normalizeContainerImageScanStatus := func(status stasv1alpha1.ContainerImageScanStatus) stasv1alpha1.ContainerImageScanStatus {
		status.LastScanTime = nil
		for i := range status.Conditions {
			status.Conditions[i].LastTransitionTime = metav1.Time{}
			status.Conditions[i].Message = "<MESSAGE>"
		}
		return status
	}

	It("should reconcile status", func() {
		cis := &stasv1alpha1.ContainerImageScan{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "nginx-unprivileged",
				Namespace: "default",
			},
			Spec: stasv1alpha1.ContainerImageScanSpec{
				ImageScanSpec: stasv1alpha1.ImageScanSpec{
					Image: stasv1alpha1.Image{
						Name:   "docker.io/nginxinc/nginx-unprivileged",
						Digest: "sha256:a96370b18b3d7e70b7b34d49dcb621a805c15cf71217ee8c77be5a98cc793fd3",
					},
				},
			},
		}
		Expect(k8sClient.Create(ctx, cis)).To(Succeed())

		// Wait for CIS to be processed by controller
		Eventually(komega.Object(cis)).Should(HaveField("Status.ObservedGeneration", Not(BeZero())))
		expectedStatus := stasv1alpha1.ContainerImageScanStatus{
			ObservedGeneration: 1,
			Conditions: []metav1.Condition{{
				Type:    "Reconciling",
				Status:  "True",
				Reason:  "ScanJobCreated",
				Message: "<MESSAGE>",
			}},
		}
		Expect(cis.Status).Should(WithTransform(normalizeContainerImageScanStatus, Equal(expectedStatus)))

	})

	It("should rescan when due", func() {
		cis := &stasv1alpha1.ContainerImageScan{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "nginx-rescan",
				Namespace: "default",
			},
			Spec: stasv1alpha1.ContainerImageScanSpec{
				ImageScanSpec: stasv1alpha1.ImageScanSpec{
					Image: stasv1alpha1.Image{
						Name:   "docker.io/nginxinc/nginx-unprivileged",
						Digest: "sha256:38c2aa106718a39d89e27b2124402ff48bdf4ed582beae62e5a3ee23d8b41f80",
					},
				},
			},
		}
		Expect(k8sClient.Create(ctx, cis)).To(Succeed())

		// Wait for scan job to be created
		scanJob := getContainerImageScanJob(cis)

		// Delete scan job and verify it's actually removed
		Expect(k8sClient.Delete(ctx, scanJob, client.PropagationPolicy(metav1.DeletePropagationOrphan))).To(Succeed())
		Expect(komega.Update(scanJob, func() {
			// Must remove finalizers to REALLY get job deleted
			scanJob.Finalizers = nil
		})()).To(Succeed())
		Eventually(komega.Get(scanJob)).Should(WithTransform(errors.ReasonForError, Equal(metav1.StatusReasonNotFound)))

		// Modify LastScanTime to trigger rescan
		Expect(komega.UpdateStatus(cis, func() {
			cis.Status.LastScanTime = &metav1.Time{Time: time.Now().Add(-time.Hour * 12)}
		})()).To(Succeed())

		// Assert new scan job created
		scanJob2 := getContainerImageScanJob(cis)
		Expect(scanJob2.UID).To(Not(Equal(scanJob.UID)))
	})

	normalizeUntestableScanJobFields := func(job *batchv1.Job) *batchv1.Job {
		job.APIVersion = "batch/v1"
		job.Kind = "Job"
		job.UID = ""
		job.ResourceVersion = ""
		job.CreationTimestamp = metav1.Time{}
		job.ManagedFields = nil
		for k := range job.Labels {
			if k == stasv1alpha1.LabelStatnettControllerUID {
				job.Labels[k] = "<CIS-UID>"
			}
		}
		job.Spec.Selector = nil
		for k := range job.Spec.Template.Labels {
			switch k {
			case "controller-uid":
				job.Spec.Template.Labels[k] = "<CONTROLLER-UID>"
			case "job-name":
				job.Spec.Template.Labels[k] = "<JOB-NAME>"
			case stasv1alpha1.LabelStatnettControllerUID:
				job.Spec.Template.Labels[k] = "<CIS-UID>"
			}

		}
		for _, container := range job.Spec.Template.Spec.Containers {
			if container.Name == trivy.ScanJobContainerName {
				for i, ev := range container.Env {
					if ev.Name == "TRIVY_TEMPLATE" {
						container.Env[i].Value = "<REPORT-TEMPLATE>"
					}
				}
			}
		}
		return job
	}

	It("should create expected scan Job", func() {
		workloadPod := &corev1.Pod{}
		Expect(yaml.FromFile(path.Join("testdata", "scan-job", "workload-pod.yaml"), workloadPod)).To(Succeed())
		Expect(k8sClient.Create(ctx, workloadPod)).To(Succeed())

		cis := &stasv1alpha1.ContainerImageScan{}
		Expect(yaml.FromFile(path.Join("testdata", "scan-job", "cis.yaml"), cis)).To(Succeed())
		Expect(controllerutil.SetOwnerReference(workloadPod, cis, k8sScheme)).To(Succeed())
		Expect(k8sClient.Create(ctx, cis)).To(Succeed())

		scanJob := getContainerImageScanJob(cis)
		expectedScanJob := &batchv1.Job{}
		Expect(yaml.FromFile(path.Join("testdata", "scan-job", "expected-scan-job.yaml"), expectedScanJob)).To(Succeed())
		Expect(scanJob).Should(WithTransform(normalizeUntestableScanJobFields, BeComparableTo(expectedScanJob)))
	})
})
