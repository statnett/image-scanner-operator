package stas

import (
	"context"

	batchv1 "k8s.io/api/batch/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	stasv1alpha1 "github.com/statnett/image-scanner-operator/api/stas/v1alpha1"
)

const (
	indexOwnerUID  = ".metadata.owner"
	indexUID       = ".metadata.uid"
	indexJobStatus = ".status.type"

	jobStatusNotFinished = ""
)

type Indexer struct{}

// SetupWithManager sets up the indexer with the Manager.
func (r *Indexer) SetupWithManager(mgr ctrl.Manager) error {
	indexer := mgr.GetFieldIndexer()

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

	jobStatusFn := func(obj client.Object) []string {
		job := obj.(*batchv1.Job)
		// TODO: non-exact field matches are not supported by the cache
		// https://github.com/kubernetes-sigs/controller-runtime/blob/main/pkg/cache/internal/cache_reader.go#L116-L121
		switch js := jobStatus(job); js {
		case batchv1.JobComplete:
			return []string{string(js)}
		case batchv1.JobFailed:
			return []string{string(js)}
		default:
			return []string{jobStatusNotFinished}
		}
	}
	for _, object := range []client.Object{&batchv1.Job{}} {
		if err := indexer.IndexField(context.TODO(), object, indexJobStatus, jobStatusFn); err != nil {
			return err
		}
	}

	return nil
}
