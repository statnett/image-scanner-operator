package v1alpha1

const (
	LabelK8sAppName                  = "app.kubernetes.io/name"
	LabelK8SAppManagedBy             = "app.kubernetes.io/managed-by"
	LabelStatnettControllerHash      = "controller.statnett.no/hash"
	LabelStatnettControllerNamespace = "controller.statnett.no/namespace"
	LabelStatnettControllerUID       = "controller.statnett.no/uid"

	AppNameImageScanner = "image-scanner"
	AppNameTrivy        = "trivy"
)
