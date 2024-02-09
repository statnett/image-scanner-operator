package controller

import (
	"context"
	"errors"
	"reflect"
	"testing"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	ctrl "sigs.k8s.io/controller-runtime"
)

func TestReconcile(t *testing.T) {
	gr := schema.GroupResource{}

	type args struct {
		ctx         context.Context
		reconcileFn ReconcileFn
	}

	tests := []struct {
		name    string
		args    args
		want    ctrl.Result
		wantErr bool
	}{
		{name: "no error", args: args{reconcileFn: reconcileError(nil)}},
		{name: "unrelated", args: args{reconcileFn: reconcileError(errors.New("error"))}, wantErr: true},
		{name: "conflict", args: args{reconcileFn: reconcileError(apierrors.NewConflict(gr, "name", nil))}, want: ctrl.Result{Requeue: true}},
		{name: "already exists", args: args{reconcileFn: reconcileError(apierrors.NewAlreadyExists(gr, "name"))}, want: ctrl.Result{Requeue: true}},
		{name: "not found", args: args{reconcileFn: reconcileError(apierrors.NewNotFound(gr, "name"))}},
		{name: "namespace terminating", args: args{reconcileFn: reconcileError(newNamespaceTerminatingError())}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Reconcile(tt.args.ctx, tt.args.reconcileFn)
			if (err != nil) != tt.wantErr {
				t.Errorf("Reconcile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Reconcile() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func reconcileError(err error) ReconcileFn {
	return func(ctx context.Context) (ctrl.Result, error) {
		return ctrl.Result{}, err
	}
}

// newNamespaceTerminatingError returns a (simplified) error indicating the namespace is terminating.
func newNamespaceTerminatingError() error {
	return &apierrors.StatusError{
		ErrStatus: metav1.Status{
			Details: &metav1.StatusDetails{
				Causes: []metav1.StatusCause{{Type: corev1.NamespaceTerminatingCause}},
			},
		}}
}
