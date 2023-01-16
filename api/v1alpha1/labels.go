package v1alpha1

const (
	LabelK8sAppName                  = "app.kubernetes.io/name"
	LabelK8SAppManagedBy             = "app.kubernetes.io/managed-by"
	LabelStatnettControllerNamespace = "controller.statnett.no/namespace"
	LabelStatnettControllerUID       = "controller.statnett.no/uid"
	LabelStatnettWorkloadKind        = "workload.statnett.no/kind"
	LabelStatnettWorkloadName        = "workload.statnett.no/name"
	LabelStatnettWorkloadNamespace   = "workload.statnett.no/namespace"

	AppNameImageScanner = "image-scanner"
	AppNameTrivy        = "trivy"
)
