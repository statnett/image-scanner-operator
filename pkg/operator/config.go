package operator

import (
	"time"

	stasv1alpha1 "github.com/statnett/image-scanner-operator/api/stas/v1alpha1"
)

type Config struct {
	MetricsLabels         []string      `mapstructure:"cis-metrics-labels"`
	ScanInterval          time.Duration `mapstructure:"scan-interval"`
	ScanJobNamespace      string        `mapstructure:"scan-job-namespace"`
	ScanJobServiceAccount string        `mapstructure:"scan-job-service-account"`
	ScanNamespaces        []string      `mapstructure:"namespaces"`
	ScanWorkloadResources []string      `mapstructure:"scan-workload-resources"`
	TrivyImage            string        `mapstructure:"trivy-image"`
	TrivyServer           string        `mapstructure:"trivy-server"`
}

func (c Config) TimeUntilNextScan(cis *stasv1alpha1.ContainerImageScan) time.Duration {
	if cis.Status.ObservedGeneration != cis.Generation || cis.Status.LastScanTime.IsZero() {
		return 0
	}

	return time.Until(cis.Status.LastScanTime.Add(c.ScanInterval))
}
