package config

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	stasv1alpha1 "github.com/statnett/image-scanner-operator/api/stas/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"
)

var _ = Describe("TimeUntilNextScan", func() {
	var config Config
	var cis *stasv1alpha1.ContainerImageScan

	BeforeEach(func() {
		config = Config{
			ScanInterval: time.Hour * 12,
		}
		cis = &stasv1alpha1.ContainerImageScan{}
	})

	It("should return zero for empty", func() {
		Expect(config.TimeUntilNextScan(cis)).To(BeZero())
	})

	Context("Recan due", func() {
		dueTime := time.Now().Add(time.Hour * -13)

		It("should NOT return zero when due", func() {
			cis.Status.LastScanTime = &metav1.Time{Time: dueTime}
			Expect(config.TimeUntilNextScan(cis)).To(BeNumerically("<", 0))
		})
	})

	Context("Recan NOT due", func() {
		notDueTime := time.Now().Add(time.Hour * -1)

		It("should return zero when generation is not observed", func() {
			cis.Status.LastScanTime = &metav1.Time{Time: notDueTime}
			cis.Generation = 1
			Expect(config.TimeUntilNextScan(cis)).To(BeZero())
		})

		It("should return zero when generation is not observed", func() {
			cis.Status.LastScanTime = &metav1.Time{Time: notDueTime}
			cis.Generation = 1
			Expect(config.TimeUntilNextScan(cis)).To(BeZero())
		})

		It("should indicate duration to next scan", func() {
			cis.Status.LastScanTime = &metav1.Time{Time: notDueTime}
			Expect(config.TimeUntilNextScan(cis)).To(BeNumerically("~", time.Hour*11, time.Second))
		})
	})
})
