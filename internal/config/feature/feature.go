package feature

import (
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/component-base/featuregate"

	"github.com/statnett/image-scanner-operator/internal/config"
)

const (
	// PolicyReport will ensure PolicyReport resources are created for completed scan jobs.
	PolicyReport featuregate.Feature = "PolicyReport"
	// NoCISStatusVulnerabilities is a feature gate to disable detailed list of vulnerabilities in ContainerImageScan status.
	NoCISStatusVulnerabilities featuregate.Feature = "NoCISStatusVulnerabilities"
)

func init() {
	runtime.Must(config.DefaultMutableFeatureGate.Add(defaultFeatureGates))
}

var defaultFeatureGates = map[featuregate.Feature]featuregate.FeatureSpec{
	PolicyReport:               {Default: true, PreRelease: featuregate.Beta},
	NoCISStatusVulnerabilities: {Default: false, PreRelease: featuregate.Alpha},
}
