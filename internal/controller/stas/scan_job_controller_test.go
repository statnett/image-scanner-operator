package stas

import (
	"bytes"
	"context"
	"io"
	"os"
	"path"
	"sort"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/mock"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/envtest/komega"

	stasv1alpha1 "github.com/statnett/image-scanner-operator/api/stas/v1alpha1"
	"github.com/statnett/image-scanner-operator/internal/trivy"
	"github.com/statnett/image-scanner-operator/internal/yaml"
)

var _ = Describe("Scan Job controller", func() {
	const (
		timeout  = 20 * time.Minute
		interval = 100 * time.Millisecond
	)

	BeforeEach(func() {
		ctx = context.Background()
	})

	Context("When scan job is complete", func() {
		It("should write scan results back to CIS status", func() {
			// Create CIS
			cis := &stasv1alpha1.ContainerImageScan{}
			Expect(yaml.FromFile(path.Join("testdata", "scan-job-successful", "successful-scan-cis.yaml"), cis)).To(Succeed())
			Expect(k8sClient.Create(ctx, cis)).To(Succeed())

			// Wait for CIS to get reconciled
			Eventually(komega.Object(cis)).Should(HaveField("Status.ObservedGeneration", Not(BeZero())))
			// Sanity check for conditions set
			Expect(cis.Status.Conditions).To(Not(BeEmpty()))

			// Simulate scan job complete
			scanJob := getContainerImageScanJob(cis)
			createScanJobPodWithLogs(scanJob, path.Join("testdata", "scan-job-successful", "successful-scan-job-pod.log.json"))
			Expect(komega.UpdateStatus(scanJob, func() {
				scanJob.Status.Succeeded = 1
			})()).To(Succeed())

			// Wait for Job to get reconciled
			Eventually(komega.Object(cis), timeout, interval).Should(HaveField("Status.LastScanTime", Not(BeZero())))
			Expect(cis.Status.LastSuccessfulScanTime).To(Not(BeZero()))
			Expect(cis.Status.LastScanJobName).To(Equal(scanJob.Name))
			// Check no conditions
			Expect(cis.Status.Conditions).To(BeEmpty())
			// Check scan results available
			Expect(cis.Status.Vulnerabilities).To(Not(BeEmpty()))
			expectedVulnSummary := &stasv1alpha1.VulnerabilitySummary{
				SeverityCount: map[string]int32{
					"CRITICAL": 4,
					"HIGH":     15,
					"MEDIUM":   0,
					"LOW":      0,
					"UNKNOWN":  0,
				},
				FixedCount:   0,
				UnfixedCount: 19,
			}
			Expect(cis.Status.VulnerabilitySummary).To(Equal(expectedVulnSummary))
		})

		Context("and scan report is too big", func() {
			It("should report correct conditions", func() {
				// Create CIS
				cis := &stasv1alpha1.ContainerImageScan{}
				Expect(yaml.FromFile(path.Join("testdata", "scan-job-successful-long", "cis.yaml"), cis)).To(Succeed())
				Expect(k8sClient.Create(ctx, cis)).To(Succeed())

				// Wait for CIS to get reconciled
				Eventually(komega.Object(cis)).Should(HaveField("Status.ObservedGeneration", Not(BeZero())))
				// Sanity check for conditions set
				Expect(cis.Status.Conditions).To(Not(BeEmpty()))

				// Simulate scan job complete
				scanJob := getContainerImageScanJob(cis)
				createScanJobPodWithLogs(scanJob, path.Join("testdata", "scan-job-successful-long", "scan-job-pod.log.json"))
				Expect(komega.UpdateStatus(scanJob, func() {
					scanJob.Status.Succeeded = 1
				})()).To(Succeed())

				// Wait for Job to get reconciled
				Eventually(komega.Object(cis), timeout, interval).Should(HaveField("Status.LastScanTime", Not(BeZero())))
				Expect(cis.Status.LastSuccessfulScanTime).To(BeZero())
				Expect(cis.Status.LastScanJobName).To(Equal(scanJob.Name))
				// Check conditions
				Expect(cis.Status.Conditions).To(HaveLen(1))
				condition := cis.Status.Conditions[0]
				Expect(condition.Type).To(Equal("Stalled"))
				Expect(condition.Status).To(Equal(metav1.ConditionTrue))
				Expect(condition.Reason).To(Equal("VulnerabilityOverflow"))
				Expect(condition.Message).To(Not(BeEmpty()))
			})
		})

		Context("but scan report is invalid JSON", func() {
			It("should report stalled condition", func() {
				// Create CIS
				cis := &stasv1alpha1.ContainerImageScan{}
				Expect(yaml.FromFile(path.Join("testdata", "scan-job-invalid-json", "cis.yaml"), cis)).To(Succeed())
				Expect(k8sClient.Create(ctx, cis)).To(Succeed())

				// Wait for CIS to get reconciled
				Eventually(komega.Object(cis)).Should(HaveField("Status.ObservedGeneration", Not(BeZero())))
				// Sanity check for conditions set
				Expect(cis.Status.Conditions).To(Not(BeEmpty()))

				// Simulate scan job complete
				scanJob := getContainerImageScanJob(cis)
				createScanJobPodWithLogs(scanJob, path.Join("testdata", "scan-job-invalid-json", "scan-job-pod.log.invalid.json"))
				Expect(komega.UpdateStatus(scanJob, func() {
					scanJob.Status.Succeeded = 1
				})()).To(Succeed())

				// Wait for Job to get reconciled
				Eventually(komega.Object(cis), timeout, interval).Should(HaveField("Status.LastScanTime", Not(BeZero())))
				Expect(cis.Status.LastSuccessfulScanTime).To(BeZero())
				Expect(cis.Status.LastScanJobName).To(Equal(scanJob.Name))
				// Check conditions
				Expect(cis.Status.Conditions).To(HaveLen(1))
				condition := cis.Status.Conditions[0]
				Expect(condition.Type).To(Equal("Stalled"))
				Expect(condition.Status).To(Equal(metav1.ConditionTrue))
				Expect(condition.Reason).To(Equal("ScanReportDecodeError"))
				Expect(condition.Message).To(Not(BeEmpty()))
			})
		})
	})

	Context("When scan job is failed", func() {
		It("should write scan results back to CIS status", func() {
			// Create CIS
			cis := &stasv1alpha1.ContainerImageScan{}
			Expect(yaml.FromFile(path.Join("testdata", "scan-job-failed", "failed-scan-cis.yaml"), cis)).To(Succeed())
			Expect(k8sClient.Create(ctx, cis)).To(Succeed())

			// Wait for CIS to get reconciled
			Eventually(komega.Object(cis)).Should(HaveField("Status.ObservedGeneration", Not(BeZero())))
			// Sanity check for conditions set
			Expect(cis.Status.Conditions).To(Not(BeEmpty()))

			// Simulate scan job complete
			scanJob := getContainerImageScanJob(cis)
			backoffLimit := *scanJob.Spec.BackoffLimit
			for i := int32(0); i < backoffLimit; i++ {
				createScanJobPodWithLogs(scanJob, path.Join("testdata", "scan-job-failed", "failed-scan-job-pod.log"))
			}
			Expect(komega.UpdateStatus(scanJob, func() {
				scanJob.Status.Failed = backoffLimit
			})()).To(Succeed())

			// Wait for Job to get reconciled
			Eventually(komega.Object(cis), timeout, interval).Should(HaveField("Status.LastScanTime", Not(BeZero())))
			Expect(cis.Status.LastSuccessfulScanTime).To(BeZero())
			Expect(cis.Status.LastScanJobName).To(Equal(scanJob.Name))
			// Check conditions
			Expect(cis.Status.Conditions).To(HaveLen(1))
			condition := cis.Status.Conditions[0]
			Expect(condition.Type).To(Equal("Stalled"))
			Expect(condition.Status).To(Equal(metav1.ConditionTrue))
			Expect(condition.Reason).To(Equal("Error"))
			Expect(condition.Message).To(Not(BeEmpty()))
		})
	})
})

var _ = DescribeTable("Converting to vulnerability summary (severity count)",
	func(vulnerabilities []stasv1alpha1.Vulnerability, expectedSummary map[string]int32) {
		summary := vulnerabilitySummary(vulnerabilities, stasv1alpha1.SeverityHigh)
		Expect(summary.SeverityCount).To(Equal(expectedSummary))
	},
	Entry("When no vulnerabilities", nil, map[string]int32{"CRITICAL": 0, "HIGH": 0}),
	Entry("When single severity", []stasv1alpha1.Vulnerability{{Severity: "CRITICAL"}}, map[string]int32{"CRITICAL": 1, "HIGH": 0}),
	Entry("When severity outside scope", []stasv1alpha1.Vulnerability{{Severity: "LOW"}}, map[string]int32{"CRITICAL": 0, "HIGH": 0, "LOW": 1}),
)

var _ = Describe("Sorting pods", func() {
	It("Should sort by startTime DESC", func() {
		var pods []corev1.Pod
		for i := 0; i < 3; i++ {
			pod := corev1.Pod{}
			now := metav1.Now()
			pod.Status.StartTime = &now
			pods = append(pods, pod)
		}

		expectedPods := make([]corev1.Pod, len(pods))
		copy(expectedPods, pods)

		sortByStartTimeDesc(pods)
		Expect(pods).NotTo(Equal(expectedPods))

		sort.SliceStable(expectedPods, func(i, j int) bool {
			return i > j
		})
		Expect(pods).To(Equal(expectedPods))
	})
})

func getContainerImageScanJob(cis *stasv1alpha1.ContainerImageScan) *batchv1.Job {
	jobs := &batchv1.JobList{}
	listOps := []client.ListOption{
		client.InNamespace(scanJobNamespace),
		client.MatchingLabels(map[string]string{stasv1alpha1.LabelStatnettControllerUID: string(cis.UID)}),
	}
	Expect(k8sClient.List(ctx, jobs, listOps...)).To(Succeed())
	Expect(jobs.Items).To(HaveLen(1))

	return &jobs.Items[0]
}

func createScanJobPodWithLogs(job *batchv1.Job, logFilePath string) {
	podLog, err := os.ReadFile(logFilePath)
	Expect(err).NotTo(HaveOccurred())

	pod := &corev1.Pod{}
	pod.Namespace = job.Namespace
	pod.GenerateName = job.Name
	pod.Labels = job.Spec.Template.Labels
	pod.Spec = job.Spec.Template.Spec
	Expect(controllerutil.SetControllerReference(job, pod, k8sScheme)).To(Succeed())
	Expect(k8sClient.Create(ctx, pod)).To(Succeed())

	logsReader.EXPECT().
		GetLogs(mock.Anything, client.ObjectKeyFromObject(pod), trivy.ScanJobContainerName).
		Call.Return(io.NopCloser(bytes.NewReader(podLog)), nil)

	now := metav1.Now()
	pod.Status.StartTime = &now
	Expect(k8sClient.Status().Update(ctx, pod)).To(Succeed())
}
