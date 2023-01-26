package metrics

import (
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/prometheus/client_golang/prometheus/testutil"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	stasv1alpha1 "github.com/statnett/image-scanner-operator/api/stas/v1alpha1"
	"github.com/statnett/image-scanner-operator/internal/config"
)

var _ = Describe("ContainerImageScan Collector", func() {
	var imageMetricsCollector *ImageMetricsCollector

	JustBeforeEach(func() {
		c := newClientWithTestdata()
		imageMetricsCollector = &ImageMetricsCollector{
			Client: c,
			Config: config.Config{MetricsLabels: []string{"system.statnett.no/name", "app.kubernetes.io/name"}},
		}
		Expect(imageMetricsCollector.SetupWithManager(&fakeManager{})).To(Succeed())
	})

	AssertNoLintIssues := func() {
		It("should not have lint issues", func() {
			problems, err := testutil.CollectAndLint(imageMetricsCollector)
			Expect(err).To(Succeed())
			Expect(problems).To(BeEmpty())
		})
	}

	Context("Trivy scanning", func() {
		AssertNoLintIssues()

		It("should produce correct success metrics", func() {
			const expected = `
			# HELP image_scanner_container_image_scan_success Displays whether or not container image scan was a success
			# TYPE image_scanner_container_image_scan_success gauge
      		image_scanner_container_image_scan_success{app_kubernetes_io_name="bad-app",image_digest="sha256:aa5f8d668258d929ee42f000b71318379a86b56e72e301ced34df8887ccbc76a",image_name="my.registry.com/my-namespace/bad-app",image_tag="latest",name="bad-app-666666",namespace="default",system_statnett_no_name="dark-side"} 1
      		image_scanner_container_image_scan_success{app_kubernetes_io_name="good-app",image_digest="sha256:293d59096e2bf7bce8c8af44086f5d3f81c98cad87928837b1cb52a61041e5d5",image_name="my.registry.com/my-namespace/good-image",image_tag="",name="good-app-aaa123",namespace="default",system_statnett_no_name="light-side"} 1
      		image_scanner_container_image_scan_success{app_kubernetes_io_name="scan-failure",image_digest="sha256:babaa4d10a7e388a37b8d41069438518184f13cdec20c580f16114b84819618b",image_name="my.registry.com/my-namespace/scan-error",image_tag="latest",name="scan-failure-111111",namespace="default",system_statnett_no_name="light-side"} 0
		`
			Expect(testutil.CollectAndCompare(imageMetricsCollector, strings.NewReader(expected), "image_scanner_container_image_scan_success")).
				To(Succeed())
		})

		It("should produce correct issues metrics", func() {
			const expected = `
			# HELP image_scanner_container_image_issues Number of container image scan issues
			# TYPE image_scanner_container_image_issues gauge
			image_scanner_container_image_issues{app_kubernetes_io_name="bad-app",image_digest="sha256:aa5f8d668258d929ee42f000b71318379a86b56e72e301ced34df8887ccbc76a",image_name="my.registry.com/my-namespace/bad-app",image_tag="latest",name="bad-app-666666",namespace="default",severity="CRITICAL",system_statnett_no_name="dark-side"} 1
			image_scanner_container_image_issues{app_kubernetes_io_name="bad-app",image_digest="sha256:aa5f8d668258d929ee42f000b71318379a86b56e72e301ced34df8887ccbc76a",image_name="my.registry.com/my-namespace/bad-app",image_tag="latest",name="bad-app-666666",namespace="default",severity="HIGH",system_statnett_no_name="dark-side"} 5
		`
			Expect(testutil.CollectAndCompare(imageMetricsCollector, strings.NewReader(expected), "image_scanner_container_image_issues")).
				To(Succeed())
		})

		It("should produce correct patch status metrics", func() {
			const expected = `
			# HELP image_scanner_container_image_patch_status Number of detected container image vulnerabilities grouped by fixed/unfixed
      		# TYPE image_scanner_container_image_patch_status gauge
      		image_scanner_container_image_patch_status{app_kubernetes_io_name="bad-app",condition="fixed",image_digest="sha256:aa5f8d668258d929ee42f000b71318379a86b56e72e301ced34df8887ccbc76a",image_name="my.registry.com/my-namespace/bad-app",image_tag="latest",name="bad-app-666666",namespace="default",system_statnett_no_name="dark-side"} 4
      		image_scanner_container_image_patch_status{app_kubernetes_io_name="bad-app",condition="unfixed",image_digest="sha256:aa5f8d668258d929ee42f000b71318379a86b56e72e301ced34df8887ccbc76a",image_name="my.registry.com/my-namespace/bad-app",image_tag="latest",name="bad-app-666666",namespace="default",system_statnett_no_name="dark-side"} 2
		`
			Expect(testutil.CollectAndCompare(imageMetricsCollector, strings.NewReader(expected), "image_scanner_container_image_patch_status")).
				To(Succeed())
		})
	})
})

func newClientWithTestdata() client.Client {
	ciss := []runtime.Object{
		&stasv1alpha1.ContainerImageScan{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "good-app-aaa123",
				Namespace: "default",
				Labels: map[string]string{
					"system.statnett.no/name": "light-side",
					"app.kubernetes.io/name":  "good-app",
				},
			},
			Spec: stasv1alpha1.ContainerImageScanSpec{
				ImageScanSpec: stasv1alpha1.ImageScanSpec{
					Image: stasv1alpha1.Image{
						Name:   "my.registry.com/my-namespace/good-image",
						Digest: "sha256:293d59096e2bf7bce8c8af44086f5d3f81c98cad87928837b1cb52a61041e5d5",
					},
				},
			},
		},
		&stasv1alpha1.ContainerImageScan{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "bad-app-666666",
				Namespace: "default",
				Labels: map[string]string{
					"system.statnett.no/name": "dark-side",
					"app.kubernetes.io/name":  "bad-app",
				},
			},
			Spec: stasv1alpha1.ContainerImageScanSpec{
				ImageScanSpec: stasv1alpha1.ImageScanSpec{
					Image: stasv1alpha1.Image{
						Name:   "my.registry.com/my-namespace/bad-app",
						Digest: "sha256:aa5f8d668258d929ee42f000b71318379a86b56e72e301ced34df8887ccbc76a",
					},
				},
				Tag: "latest",
			},
			Status: stasv1alpha1.ContainerImageScanStatus{
				VulnerabilitySummary: &stasv1alpha1.VulnerabilitySummary{
					SeverityCount: map[string]int32{"CRITICAL": 1, "HIGH": 5},
					FixedCount:    4,
					UnfixedCount:  2,
				},
			},
		},
		&stasv1alpha1.ContainerImageScan{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "scan-failure-111111",
				Namespace: "default",
				Labels: map[string]string{
					"system.statnett.no/name": "light-side",
					"app.kubernetes.io/name":  "scan-failure",
				},
			},
			Spec: stasv1alpha1.ContainerImageScanSpec{
				ImageScanSpec: stasv1alpha1.ImageScanSpec{
					Image: stasv1alpha1.Image{
						Name:   "my.registry.com/my-namespace/scan-error",
						Digest: "sha256:babaa4d10a7e388a37b8d41069438518184f13cdec20c580f16114b84819618b",
					},
				},
				Tag: "latest",
			},
			Status: stasv1alpha1.ContainerImageScanStatus{
				Conditions: []metav1.Condition{{
					Type:    "Stalled",
					Status:  metav1.ConditionTrue,
					Reason:  "Error",
					Message: "Some error message",
				}},
			},
		},
	}
	scheme := runtime.NewScheme()
	Expect(stasv1alpha1.AddToScheme(scheme)).To(Succeed())

	return fake.NewClientBuilder().
		WithScheme(scheme).
		WithRuntimeObjects(ciss...).
		Build()
}
