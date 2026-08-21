package controller

import (
	"context"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	ctrl "sigs.k8s.io/controller-runtime"

	staserrors "github.com/statnett/image-scanner-operator/internal/errors"
)

type ReconcileFn func(context.Context) (ctrl.Result, error)

func Reconcile(ctx context.Context, reconcileFn ReconcileFn) (ctrl.Result, error) {
	result, err := reconcileFn(ctx)
	return result, staserrors.IgnoreAny(err, staserrors.IsNamespaceTerminating, apierrors.IsNotFound)
}
