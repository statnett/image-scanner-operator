package stas

import (
	"context"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	stasv1alpha1 "github.com/statnett/image-scanner-operator/api/stas/v1alpha1"
)

const (
	indexControllerUID = ".metadata.controller"
	indexOwnerUID      = ".metadata.owner"
	indexUID           = ".metadata.uid"
	indexJobCondition  = ".status.condition"

	jobNotFinished = "NotFinished"
)

type Indexer struct{}

// SetupWithManager sets up the indexer with the Manager.
func (r *Indexer) SetupWithManager(mgr ctrl.Manager) error {
	indexer := mgr.GetFieldIndexer()

	controllerUIDFn := func(obj client.Object) []string {
		controller := metav1.GetControllerOfNoCopy(obj)
		if controller == nil {
			return nil
		}

		return []string{string(controller.UID)}
	}
	for _, object := range []client.Object{&corev1.Pod{}} {
		if err := indexer.IndexField(context.TODO(), object, indexControllerUID, controllerUIDFn); err != nil {
			return err
		}
	}

	ownerUIDFn := func(obj client.Object) []string {
		ownerReferences := obj.GetOwnerReferences()

		uids := make([]string, len(ownerReferences))
		for i, or := range ownerReferences {
			uids[i] = string(or.UID)
		}

		return uids
	}
	for _, object := range []client.Object{&stasv1alpha1.ContainerImageScan{}} {
		if err := indexer.IndexField(context.TODO(), object, indexOwnerUID, ownerUIDFn); err != nil {
			return err
		}
	}

	UIDFn := func(obj client.Object) []string {
		return []string{string(obj.GetUID())}
	}
	for _, object := range []client.Object{&stasv1alpha1.ContainerImageScan{}} {
		if err := indexer.IndexField(context.TODO(), object, indexUID, UIDFn); err != nil {
			return err
		}
	}

	jobConditionFn := func(obj client.Object) []string {
		job := obj.(*batchv1.Job)
		// TODO: non-exact field matches are not supported by the cache
		// https://github.com/kubernetes-sigs/controller-runtime/blob/main/pkg/cache/internal/cache_reader.go#L116-L121
		// So mapping to [Complete, Failed, NotFinished] where the last is a composite condition
		switch jc := jobCondition(job); jc {
		case batchv1.JobComplete:
			return []string{string(jc)}
		case batchv1.JobFailed:
			return []string{string(jc)}
		default:
			return []string{jobNotFinished}
		}
	}
	for _, object := range []client.Object{&batchv1.Job{}} {
		if err := indexer.IndexField(context.TODO(), object, indexJobCondition, jobConditionFn); err != nil {
			return err
		}
	}

	return nil
}
