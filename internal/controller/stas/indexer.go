package stas

import (
	"context"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	stasv1alpha1 "github.com/statnett/image-scanner-operator/api/stas/v1alpha1"
)

const (
	indexOwnerUID = ".metadata.owner"
	indexUID      = ".metadata.uid"
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

	return nil
}
