package operator

import (
	"time"

	stasv1alpha1 "github.com/statnett/image-scanner-operator/api/v1alpha1"
)

type Config struct {
	ScanInterval          time.Duration `mapstructure:"scan-interval"`
	ScanJobNamespace      string        `mapstructure:"scan-job-namespace"`
	ScanJobServiceAccount string        `mapstructure:"scan-job-service-account"`
	ScanWorkloadResources []string      `mapstructure:"scan-workload-resources"`
	TrivyImage            string        `mapstructure:"trivy-image"`
	TrivyServer           string        `mapstructure:"trivy-server"`
}

func (c Config) TimeUntilNextScan(cis *stasv1alpha1.ContainerImageScan) time.Duration {
	if cis.Status.ObservedGeneration != cis.Generation {
		return 0
	}
	if cis.Status.LastScanTime.IsZero() {
		return 0
	}
	nextScanTime := cis.Status.LastScanTime.Add(c.ScanInterval)
	return time.Until(nextScanTime)
}
