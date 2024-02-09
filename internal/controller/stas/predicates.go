package stas

import (
	"regexp"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	eventsv1 "k8s.io/api/events/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	stasv1alpha1 "github.com/statnett/image-scanner-operator/api/stas/v1alpha1"
)

func namespaceMatchRegexp(re *regexp.Regexp) predicate.Predicate {
	return predicate.NewPredicateFuncs(func(object client.Object) bool {
		return re.MatchString(object.GetNamespace())
	})
}

func podContainerStatusImagesChanged() predicate.Predicate {
	return predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool {
			pod := e.Object.(*corev1.Pod)
			return len(pod.Status.ContainerStatuses) > 0
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			pod1 := e.ObjectOld.(*corev1.Pod)
			pod2 := e.ObjectNew.(*corev1.Pod)
			if len(pod1.Status.ContainerStatuses) != len(pod2.Status.ContainerStatuses) {
				return true
			}
			for i := 0; i < len(pod1.Status.ContainerStatuses); i++ {
				cs1 := pod1.Status.ContainerStatuses[i]
				cs2 := pod2.Status.ContainerStatuses[i]
				if cs1.Image != cs2.Image {
					return true
				}
				if cs1.ImageID != cs2.ImageID {
					return true
				}
			}
			return false
		},
	}
}

func ignoreCreationPredicate() predicate.Predicate {
	return predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool {
			return false
		},
	}
}

func ignoreDeletionPredicate() predicate.Predicate {
	return predicate.Funcs{
		DeleteFunc: func(e event.DeleteEvent) bool {
			return false
		},
	}
}

var noController = predicate.NewPredicateFuncs(func(object client.Object) bool {
	return metav1.GetControllerOf(object) == nil
})

func controllerInKinds(groupKinds ...schema.GroupKind) predicate.Predicate {
	return predicate.NewPredicateFuncs(func(object client.Object) bool {
		controller := metav1.GetControllerOf(object)
		if controller != nil {
			controllerGV, err := schema.ParseGroupVersion(controller.APIVersion)
			if err != nil {
				return false
			}

			for _, groupKind := range groupKinds {
				if controller.Kind == groupKind.Kind && controllerGV.Group == groupKind.Group {
					return true
				}
			}
		}

		return false
	})
}

func inNamespacePredicate(namespace string) predicate.Predicate {
	return predicate.NewPredicateFuncs(func(object client.Object) bool {
		return object.GetNamespace() == namespace
	})
}

var managedByImageScanner = predicate.NewPredicateFuncs(func(object client.Object) bool {
	if managedBy, ok := object.GetLabels()[stasv1alpha1.LabelK8SAppManagedBy]; ok {
		return managedBy == stasv1alpha1.AppNameImageScanner
	}
	return false
})

var jobIsFinished = predicate.NewPredicateFuncs(func(object client.Object) bool {
	return isJobFinished(object.(*batchv1.Job))
})

var cisVulnerabilityOverflow = predicate.NewPredicateFuncs(func(object client.Object) bool {
	return object.(*stasv1alpha1.ContainerImageScan).HasVulnerabilityOverflow()
})

// isJobFinished checks whether the given Job has finished execution.
// It does not discriminate between successful and failed terminations.
// https://github.com/kubernetes/kubernetes/blob/master/pkg/controller/job/utils.go#L24-L33
func isJobFinished(j *batchv1.Job) bool {
	c := jobCondition(j)
	return c == batchv1.JobComplete || c == batchv1.JobFailed
}

func jobCondition(j *batchv1.Job) batchv1.JobConditionType {
	for _, c := range j.Status.Conditions {
		if c.Status == corev1.ConditionTrue {
			return c.Type
		}
	}

	return ""
}

func eventRegardingKind(kind string) predicate.Predicate {
	return predicate.NewPredicateFuncs(func(object client.Object) bool {
		e := object.(*eventsv1.Event)
		return e.Regarding.Kind == kind
	})
}

func eventReason(reason string) predicate.Predicate {
	return predicate.NewPredicateFuncs(func(object client.Object) bool {
		e := object.(*eventsv1.Event)
		return e.Reason == reason
	})
}
