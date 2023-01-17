package stas

import (
	"context"
	"path"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/envtest/komega"

	"github.com/statnett/image-scanner-operator/api/stas/v1alpha1"
	"github.com/statnett/image-scanner-operator/internal/trivy"
	"github.com/statnett/image-scanner-operator/internal/yaml"
)

var _ = Describe("ContainerImageScan controller", func() {
	BeforeEach(func() {
		ctx = context.Background()
	})

	normalizeContainerImageScanStatus := func(status v1alpha1.ContainerImageScanStatus) v1alpha1.ContainerImageScanStatus {
		status.LastScanTime = nil
		for i := range status.Conditions {
			status.Conditions[i].LastTransitionTime = metav1.Time{}
			status.Conditions[i].Message = "<MESSAGE>"
		}
		return status
	}

	It("should reconcile status", func() {
		cis := &v1alpha1.ContainerImageScan{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "nginx-unprivileged",
				Namespace: "default",
			},
			Spec: v1alpha1.ContainerImageScanSpec{
				ImageScanSpec: v1alpha1.ImageScanSpec{
					Image: v1alpha1.Image{
						Name:   "docker.io/nginxinc/nginx-unprivileged",
						Digest: "sha256:a96370b18b3d7e70b7b34d49dcb621a805c15cf71217ee8c77be5a98cc793fd3",
					},
				},
			},
		}
		Expect(k8sClient.Create(ctx, cis)).To(Succeed())

		// Wait for CIS to be processed by controller
		Eventually(komega.Object(cis)).Should(HaveField("Status.ObservedGeneration", Not(BeZero())))
		expectedStatus := v1alpha1.ContainerImageScanStatus{
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

	normalizeUntestableScanJobFields := func(job *batchv1.Job) *batchv1.Job {
		job.APIVersion = "batch/v1"
		job.Kind = "Job"
		job.UID = ""
		job.ResourceVersion = ""
		job.CreationTimestamp = metav1.Time{}
		job.ManagedFields = nil
		for k := range job.Labels {
			if k == v1alpha1.LabelStatnettControllerUID {
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

		cis := &v1alpha1.ContainerImageScan{}
		Expect(yaml.FromFile(path.Join("testdata", "scan-job", "cis.yaml"), cis)).To(Succeed())
		Expect(controllerutil.SetOwnerReference(workloadPod, cis, k8sScheme)).To(Succeed())
		Expect(k8sClient.Create(ctx, cis)).To(Succeed())

		jobs := &batchv1.JobList{}
		listOps := []client.ListOption{
			client.InNamespace(scanJobNamespace),
			client.MatchingLabels(map[string]string{v1alpha1.LabelStatnettControllerUID: string(cis.UID)}),
		}
		Eventually(komega.ObjectList(jobs, listOps...)).Should(HaveField("Items", HaveLen(1)))
		scanJob := &jobs.Items[0]

		expectedScanJob := &batchv1.Job{}
		Expect(yaml.FromFile(path.Join("testdata", "scan-job", "expected-scan-job.yaml"), expectedScanJob)).To(Succeed())
		Expect(scanJob).Should(WithTransform(normalizeUntestableScanJobFields, BeComparableTo(expectedScanJob)))
	})
})
