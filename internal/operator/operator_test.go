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
		fs = flag.NewFlagSet("read from env", flag.ExitOnError)
		cfg = &config.Config{}
		opr.BindFlags(cfg, fs)
	})

	Context("Using scan-namespace-exclude-regexp flag", func() {
		It("Should have correct default", func() {
			fs.Parse(nil)
			opr.UnmarshalConfig(cfg)
			Expect(cfg.ScanNamespaceExcludeRegexp).NotTo(BeNil())
			Expect(cfg.ScanNamespaceExcludeRegexp.String()).To(Equal("^(kube-|openshift-).*"))
		})

		It("Should be configurable", func() {
			fs.Parse([]string{"--scan-namespace-exclude-regexp=^$"})
			opr.UnmarshalConfig(cfg)
			Expect(cfg.ScanNamespaceExcludeRegexp).NotTo(BeNil())
			Expect(cfg.ScanNamespaceExcludeRegexp.String()).To(Equal("^$"))
		})
	})

})
