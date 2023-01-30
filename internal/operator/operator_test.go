package operator

import (
	"flag"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/statnett/image-scanner-operator/internal/config"
)

var _ = Describe("Operator config from flags", func() {
	var (
		opr Operator = Operator{}
		fs  *flag.FlagSet
		cfg *config.Config
	)

	BeforeEach(func() {
		fs = flag.NewFlagSet("test", flag.ExitOnError)
		cfg = &config.Config{}
		Expect(opr.BindFlags(cfg, fs)).To(Succeed())
	})

	Context("Using scan-namespace-exclude-regexp flag", func() {
		It("Should have correct default", func() {
			Expect(fs.Parse(nil)).To(Succeed())
			Expect(opr.UnmarshalConfig(cfg)).To(Succeed())
			Expect(cfg.ScanNamespaceExcludeRegexp).NotTo(BeNil())
			Expect(cfg.ScanNamespaceExcludeRegexp.String()).To(Equal("^(kube-|openshift-).*"))
		})

		It("Should be configurable", func() {
			args := []string{"--scan-namespace-exclude-regexp=^$"}
			Expect(fs.Parse(args)).To(Succeed())
			Expect(opr.UnmarshalConfig(cfg)).To(Succeed())
			Expect(cfg.ScanNamespaceExcludeRegexp).NotTo(BeNil())
			Expect(cfg.ScanNamespaceExcludeRegexp.String()).To(Equal("^$"))
		})

		It("Should error on invalid regexp", func() {
			args := []string{"--scan-namespace-exclude-regexp=["}
			Expect(fs.Parse(args)).To(Succeed())
			Expect(opr.UnmarshalConfig(cfg)).To(MatchError(ContainSubstring("error parsing regexp")))
		})
	})

})
