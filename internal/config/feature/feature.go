package feature

import (
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/component-base/featuregate"

	"github.com/statnett/image-scanner-operator/internal/config"
)

// PolicyReport will ensure PolicyReport resources are created for completed scan jobs.
const PolicyReport featuregate.Feature = "PolicyReport"

func init() {
	runtime.Must(config.DefaultMutableFeatureGate.Add(defaultFeatureGates))
}

var defaultFeatureGates = map[featuregate.Feature]featuregate.FeatureSpec{
	PolicyReport: {Default: true, PreRelease: featuregate.Beta},
}
