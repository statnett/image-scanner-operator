package errors

import (
	"errors"
	"testing"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestIgnoreAny(t *testing.T) {
	err := errors.New("error")

	type args struct {
		err error
		is  []ErrorIs
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{name: "ignored", args: args{err: err, is: []ErrorIs{isErrorAlways}}},
		{name: "not ignored", args: args{err: err, is: []ErrorIs{isErrorNever}}, wantErr: true},
		{name: "ignored wins 1", args: args{err: err, is: []ErrorIs{isErrorAlways, isErrorNever}}},
		{name: "ignored wins 2", args: args{err: err, is: []ErrorIs{isErrorNever, isErrorAlways}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := IgnoreAny(tt.args.err, tt.args.is...); (err != nil) != tt.wantErr {
				t.Errorf("IgnoreAny() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestIgnore(t *testing.T) {
	err := errors.New("error")

	type args struct {
		err error
		is  ErrorIs
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{name: "ignored", args: args{err: err, is: isErrorAlways}},
		{name: "not ignored", args: args{err: err, is: isErrorNever}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Ignore(tt.args.err, tt.args.is); (err != nil) != tt.wantErr {
				t.Errorf("Ignore() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestIsNamespaceTerminating(t *testing.T) {
	type args struct {
		err error
	}

	tests := []struct {
		name string
		args args
		want bool
	}{
		{name: "nil", args: args{err: nil}},
		{name: "namespace terminating 1", args: args{err: newNamespaceTerminatingError()}, want: true},
		{name: "conflict update", args: args{err: errors.New("the object has been modified; please apply your changes to the latest version and try again")}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsNamespaceTerminating(tt.args.err); got != tt.want {
				t.Errorf("IsNamespaceTerminating() = %v, want %v", got, tt.want)
			}
		})
	}
}

//goland:noinspection GoUnusedParameter
func isErrorAlways(err error) bool {
	return true
}

//goland:noinspection GoUnusedParameter
func isErrorNever(err error) bool {
	return false
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
