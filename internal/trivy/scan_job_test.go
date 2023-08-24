package trivy

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/utils/ptr"

	stasv1alpha1 "github.com/statnett/image-scanner-operator/api/stas/v1alpha1"
)

var _ = Describe("Creating scan Job container", func() {
	var jobBuilder *filesystemScanJobBuilder
	var cisSpec stasv1alpha1.ContainerImageScanSpec

	BeforeEach(func() {
		jobBuilder = &filesystemScanJobBuilder{}
		cisSpec = stasv1alpha1.ContainerImageScanSpec{}
		cisSpec.Image.Name = "foo.registry/bar"
		cisSpec.Image.Digest = "sha256:f1645ab5fbbbcf9e3484d1506dd65fc9fb26dd6817cb3a0a64249d8a8973e170"
	})

	Context("minimum severity config", func() {
		It("should not include severity when minSeverity omitted", func() {
			container, err := jobBuilder.container(cisSpec)
			Expect(err).ToNot(HaveOccurred())
			Expect(container.Env).To(Not(ContainElement(HaveField("Name", Equal("TRIVY_SEVERITY")))))
		})

		It("should set severity when minSeverity set", func() {
			cisSpec.ScanConfig.MinSeverity = ptr.To("MEDIUM")
			container, err := jobBuilder.container(cisSpec)
			Expect(err).ToNot(HaveOccurred())
			expectedSeverityEnv := corev1.EnvVar{
				Name:  "TRIVY_SEVERITY",
				Value: "MEDIUM,HIGH,CRITICAL",
			}
			Expect(container.Env).To(ContainElement(expectedSeverityEnv))
		})

		It("should set ignore-unfixed when ignoreUnfixed set", func() {
			cisSpec.ScanConfig.IgnoreUnfixed = ptr.To(true)
			container, err := jobBuilder.container(cisSpec)
			Expect(err).ToNot(HaveOccurred())
			expectedSeverityEnv := corev1.EnvVar{
				Name:  "TRIVY_IGNORE_UNFIXED",
				Value: "true",
			}
			Expect(container.Env).To(ContainElement(expectedSeverityEnv))
		})
	})
})
