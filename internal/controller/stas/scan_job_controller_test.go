package stas

import (
	"bytes"
	"context"
	"io"
	"os"
	"path"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	openreportsv1alpha1 "github.com/openreports/reports-api/apis/openreports.io/v1alpha1"
	"github.com/stretchr/testify/mock"
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

var _ = Describe("Scan Job controller", func() {
	const (
		timeout  = 20 * time.Minute
		interval = 100 * time.Millisecond
	)

	BeforeEach(func() {
		ctx = context.Background()
	})

	Context("When scan job is complete", func() {
		It("should write scan results back to CIS status and create policy report", func() {
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
				setJobComplete(scanJob)
			})()).To(Succeed())

			// Wait for Job to get reconciled
			Eventually(komega.Object(cis), timeout, interval).Should(HaveField("Status.LastScanTime", Not(BeZero())))
			Expect(cis.Status.LastSuccessfulScanTime).To(Not(BeZero()))
			Expect(cis.Status.LastScanJobUID).To(Equal(scanJob.UID))
			// Check no conditions
			Expect(cis.Status.Conditions).To(BeEmpty())
			// Check scan results available
			Expect(cis.Status.Vulnerabilities).To(BeEmpty())
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

			// Check policy report exists with expected content
			report := &openreportsv1alpha1.Report{}
			report.Name = cis.Name
			report.Namespace = cis.Namespace
			Expect(komega.Get(report)()).To(Succeed())
			Expect(report.Results).To(Not(BeEmpty()))
			Expect(report.Results).Should(HaveEach(
				WithTransform(func(vulnerability openreportsv1alpha1.ReportResult) map[string]string {
					return vulnerability.Properties
				},
					Not(BeEmpty()),
				),
			))
			expectedSummary := openreportsv1alpha1.ReportSummary{
				Fail: 19,
			}
			Expect(report.Summary).To(Equal(expectedSummary))
		})

		Context("and scan report is too big", func() {
			It("should filter report by minimum severity and create policy report", func() {
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
					setJobComplete(scanJob)
				})()).To(Succeed())

				// Wait for Job to get reconciled
				Eventually(komega.Object(cis), timeout, interval).Should(HaveField("Status.LastScanTime", Not(BeZero())))
				Expect(cis.Status.LastSuccessfulScanTime).To(Not(BeZero()))
				Expect(cis.Status.LastScanJobUID).To(Equal(scanJob.UID))
				// Check no conditions
				Expect(cis.Status.Conditions).To(BeEmpty())
				// Check scan results available and filtered
				Expect(cis.Status.Vulnerabilities).To(BeEmpty())
				expectedVulnSummary := &stasv1alpha1.VulnerabilitySummary{
					SeverityCount: map[string]int32{
						"CRITICAL": 653,
						"HIGH":     2606,
						"MEDIUM":   3881,
						"LOW":      3178,
						"UNKNOWN":  77,
					},
					FixedCount:   5267,
					UnfixedCount: 5128,
				}
				Expect(cis.Status.VulnerabilitySummary).To(Equal(expectedVulnSummary))

				// Check policy report exists with expected content
				report := &openreportsv1alpha1.Report{}
				report.Name = cis.Name
				report.Namespace = cis.Namespace
				Expect(komega.Get(report)()).To(Succeed())
				Expect(report.Results).To(Not(BeEmpty()))
				Expect(report.Results).Should(HaveEach(
					WithTransform(func(vulnerability openreportsv1alpha1.ReportResult) string {
						return string(vulnerability.Severity)
					},
						SatisfyAny(
							Equal("critical"),
							Equal("high"),
						),
					),
				))
				expectedSummary := openreportsv1alpha1.ReportSummary{
					Fail: 3259,
					Warn: 7059,
					Skip: 77,
				}
				Expect(report.Summary).To(Equal(expectedSummary))
			})
		})

		Context("but scan report is invalid JSON and NOT create policy report", func() {
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
					setJobComplete(scanJob)
				})()).To(Succeed())

				// Wait for Job to get reconciled
				Eventually(komega.Object(cis), timeout, interval).Should(HaveField("Status.LastScanTime", Not(BeZero())))
				Expect(cis.Status.LastSuccessfulScanTime).To(BeZero())
				Expect(cis.Status.LastScanJobUID).To(Equal(scanJob.UID))
				// Check conditions
				Expect(cis.Status.Conditions).To(HaveLen(1))
				condition := cis.Status.Conditions[0]
				Expect(condition.Type).To(Equal("Stalled"))
				Expect(condition.Status).To(Equal(metav1.ConditionTrue))
				Expect(condition.Reason).To(Equal("ScanReportDecodeError"))
				Expect(condition.Message).To(Not(BeEmpty()))

				// Check policy report does NOT exist
				report := &openreportsv1alpha1.Report{}
				report.Name = cis.Name
				report.Namespace = cis.Namespace
				Expect(komega.Get(report)()).Should(WithTransform(errors.ReasonForError, Equal(metav1.StatusReasonNotFound)))
			})
		})
	})

	Context("When scan job is failed", func() {
		It("should write scan results back to CIS status and NOT create policy report", func() {
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
			createScanJobPodWithLogs(scanJob, path.Join("testdata", "scan-job-failed", "failed-scan-job-pod.log"))
			Expect(komega.UpdateStatus(scanJob, func() {
				setJobFailed(scanJob)
			})()).To(Succeed())

			// Wait for Job to get reconciled
			Eventually(komega.Object(cis), timeout, interval).Should(HaveField("Status.LastScanTime", Not(BeZero())))
			Expect(cis.Status.LastSuccessfulScanTime).To(BeZero())
			Expect(cis.Status.LastScanJobUID).To(Equal(scanJob.UID))
			// Check conditions
			Expect(cis.Status.Conditions).To(HaveLen(1))
			condition := cis.Status.Conditions[0]
			Expect(condition.Type).To(Equal("Stalled"))
			Expect(condition.Status).To(Equal(metav1.ConditionTrue))
			Expect(condition.Reason).To(Equal("Error"))
			Expect(condition.Message).To(Not(BeEmpty()))

			// Check policy report does NOT exist
			report := &openreportsv1alpha1.Report{}
			report.Name = cis.Name
			report.Namespace = cis.Namespace
			Expect(komega.Get(report)()).Should(WithTransform(errors.ReasonForError, Equal(metav1.StatusReasonNotFound)))
		})
	})
})

func setJobComplete(job *batchv1.Job) {
	job.Status.StartTime = &metav1.Time{Time: time.Now()}
	job.Status.CompletionTime = &metav1.Time{Time: time.Now()}
	job.Status.Conditions = []batchv1.JobCondition{
		{
			Type:   batchv1.JobSuccessCriteriaMet,
			Status: corev1.ConditionTrue,
		},
		{
			Type:   batchv1.JobComplete,
			Status: corev1.ConditionTrue,
		},
	}
}

func setJobFailed(job *batchv1.Job) {
	job.Status.StartTime = &metav1.Time{Time: time.Now()}
	job.Status.Conditions = []batchv1.JobCondition{
		{
			Type:   batchv1.JobFailureTarget,
			Status: corev1.ConditionTrue,
		},
		{
			Type:   batchv1.JobFailed,
			Status: corev1.ConditionTrue,
		},
	}
}

var _ = DescribeTable("Converting to vulnerability summary (severity count)",
	func(vulnerabilities []stasv1alpha1.Vulnerability, expectedSummary map[string]int32) {
		summary := vulnerabilitySummary(vulnerabilities, stasv1alpha1.SeverityHigh)
		Expect(summary.SeverityCount).To(Equal(expectedSummary))
	},
	Entry("When no vulnerabilities", nil, map[string]int32{"CRITICAL": 0, "HIGH": 0}),
	Entry("When single severity", []stasv1alpha1.Vulnerability{{Severity: stasv1alpha1.SeverityCritical}}, map[string]int32{"CRITICAL": 1, "HIGH": 0}),
	Entry("When severity outside scope", []stasv1alpha1.Vulnerability{{Severity: stasv1alpha1.SeverityLow}}, map[string]int32{"CRITICAL": 0, "HIGH": 0, "LOW": 1}),
)

func getContainerImageScanJob(cis *stasv1alpha1.ContainerImageScan) *batchv1.Job {
	jobs := &batchv1.JobList{}
	listOps := []client.ListOption{
		client.InNamespace(scanJobNamespace),
		client.MatchingLabels(map[string]string{stasv1alpha1.LabelStatnettControllerUID: string(cis.UID)}),
	}
	Eventually(komega.ObjectList(jobs, listOps...)).Should(HaveField("Items", HaveLen(1)))

	return &jobs.Items[0]
}

func assertNoContainerImageScanJob(cis *stasv1alpha1.ContainerImageScan) {
	jobs := &batchv1.JobList{}
	listOps := []client.ListOption{
		client.InNamespace(scanJobNamespace),
		client.MatchingLabels(map[string]string{stasv1alpha1.LabelStatnettControllerUID: string(cis.UID)}),
	}
	Consistently(komega.ObjectList(jobs, listOps...)).Should(HaveField("Items", HaveLen(0)))
}

func createScanJobPodWithLogs(job *batchv1.Job, logFilePath string) {
	podLog, err := os.ReadFile(filepath.Clean(logFilePath))
	Expect(err).NotTo(HaveOccurred())

	pod := createScanJobPod(job)

	logsReader.EXPECT().
		GetLogs(mock.Anything, client.ObjectKeyFromObject(pod), trivy.ScanJobContainerName).
		Call.Return(io.NopCloser(bytes.NewReader(podLog)), nil)
}

func createScanJobPod(job *batchv1.Job) *corev1.Pod {
	pod := &corev1.Pod{}
	pod.Namespace = job.Namespace
	pod.GenerateName = job.Name
	pod.Labels = job.Spec.Template.Labels
	pod.Spec = job.Spec.Template.Spec
	Expect(controllerutil.SetControllerReference(job, pod, k8sScheme)).To(Succeed())
	Expect(k8sClient.Create(ctx, pod)).To(Succeed())

	return pod
}
