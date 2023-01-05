package controller

import (
	"context"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	staserrors "github.com/statnett/image-scanner-operator/internal/errors"
)

type ReconcileFn func(context.Context) (ctrl.Result, error)

func Reconcile(ctx context.Context, reconcileFn ReconcileFn) (ctrl.Result, error) {
	result, err := reconcileFn(ctx)
	if apierrors.IsConflict(err) {
		// Resource conflict; requeue the request
		return ctrl.Result{Requeue: true}, nil
	}
	if apierrors.IsAlreadyExists(err) {
		// Log error message as warning (-1 = WARN)
		logf.FromContext(ctx, "error", err.Error()).
			V(-1).
			Info("Assuming transient error (race condition), requeuing request")
		return ctrl.Result{Requeue: true}, nil
	}
	return result, staserrors.IgnoreAny(err, staserrors.IsNamespaceTerminating, apierrors.IsNotFound)
}
