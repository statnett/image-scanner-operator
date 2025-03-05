package stas

import (
	"regexp"
	"slices"

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
	job := object.(*batchv1.Job)
	return isJobComplete(job) || isJobFailed(job)
})

func isJobComplete(j *batchv1.Job) bool {
	return slices.ContainsFunc(j.Status.Conditions, func(condition batchv1.JobCondition) bool {
		return condition.Type == batchv1.JobComplete && condition.Status == corev1.ConditionTrue
	})
}

func isJobFailed(j *batchv1.Job) bool {
	return slices.ContainsFunc(j.Status.Conditions, func(condition batchv1.JobCondition) bool {
		return condition.Type == batchv1.JobFailed && condition.Status == corev1.ConditionTrue
	})
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
