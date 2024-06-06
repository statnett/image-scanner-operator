package stas

import (
	"context"
	"encoding/json"
	"fmt"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"
	metav1ac "k8s.io/client-go/applyconfigurations/meta/v1"
	"k8s.io/client-go/util/csaupgrade"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"

	ctrlerrors "github.com/statnett/image-scanner-operator/internal/errors"
)

const (
	fieldOwner = client.FieldOwner("image-scanner-operator")

	// crRegressionFieldManager is the field manager that was introduced by a regression in controller-runtime
	// version 0.15.0; fixed in 15.1 and 0.16.0: https://github.com/kubernetes-sigs/controller-runtime/pull/2435
	crRegressionFieldManager = "Go-http-client"

	// beforeFirstApplyFieldManager seems to be a manager set when managedFields got introduced?
	// Or, ref. apelisse:  I can't remember, but I think at some point we didn't track managedFields until the object had been applied at least once.
	// And we put all the changes that happened before that first apply under that manager.
	beforeFirstApplyFieldManager = "before-first-apply"
)

type applyPatch struct {
	// must use any type until apply configurations implements a common interface
	patch any
}

func (p applyPatch) Type() types.PatchType {
	return types.ApplyPatchType
}

func (p applyPatch) Data(_ client.Object) ([]byte, error) {
	return json.Marshal(p.patch)
}

// FieldValidationStrict instructs the server on how to handle
// objects in the request (POST/PUT/PATCH) containing unknown
// or duplicate fields. This will fail the request with a BadRequest
// error if any unknown fields would be dropped from the object, or if any
// duplicate fields are present. The error returned from the server
// will contain all unknown and duplicate fields encountered.
var FieldValidationStrict = fieldValidationStrict{}

var (
	_ client.PatchOption            = fieldValidationStrict{}
	_ client.SubResourcePatchOption = fieldValidationStrict{}
)

type fieldValidationStrict struct{}

func (fieldValidationStrict) ApplyToPatch(opts *client.PatchOptions) {
	if opts.Raw == nil {
		opts.Raw = &metav1.PatchOptions{}
	}

	opts.Raw.FieldValidation = "Strict"
}

func (fieldValidationStrict) ApplyToSubResourcePatch(opts *client.SubResourcePatchOptions) {
	if opts.Raw == nil {
		opts.Raw = &metav1.PatchOptions{}
	}

	opts.Raw.FieldValidation = "Strict"
}

// SetOwnerReference is a helper method to make sure the given object contains an object reference to the object provided.
// This allows you to declare that owner has a dependency on the object without specifying it as a controller.
// If a reference to the same object already exists, it'll be overwritten with the newly provided version.
func SetOwnerReference(owner metav1.Object, owned *metav1ac.ObjectMetaApplyConfiguration, scheme *runtime.Scheme) error {
	// Validate the owner.
	ro, ok := owner.(runtime.Object)
	if !ok {
		return fmt.Errorf("%T is not a runtime.Object, cannot call SetOwnerReference", owner)
	}

	if err := validateOwner(owner, owned); err != nil {
		return err
	}

	gvk, err := apiutil.GVKForObject(ro, scheme)
	if err != nil {
		return err
	}

	owned.WithOwnerReferences(
		metav1ac.OwnerReference().
			WithAPIVersion(gvk.GroupVersion().String()).
			WithKind(gvk.Kind).
			WithName(owner.GetName()).
			WithUID(owner.GetUID()),
	)

	return nil
}

func validateOwner(owner metav1.Object, object *metav1ac.ObjectMetaApplyConfiguration) error {
	ownerNs := owner.GetNamespace()
	if ownerNs != "" {
		objNs := ptr.Deref(object.Namespace, "")
		if objNs == "" {
			return fmt.Errorf("cluster-scoped resource must not have a namespace-scoped owner, owner's namespace %s", ownerNs)
		}

		if ownerNs != objNs {
			return fmt.Errorf("cross-namespace owner references are disallowed, owner's namespace %s, obj's namespace %s", owner.GetNamespace(), objNs)
		}
	}

	return nil
}

func upgradeManagedFields(ctx context.Context, r client.Client, obj client.Object, fieldOwner client.FieldOwner, opts ...csaupgrade.Option) error {
	if err := r.Get(ctx, client.ObjectKeyFromObject(obj), obj); err != nil {
		// If not found, there is nothing to patch
		return ctrlerrors.Ignore(err, errors.IsNotFound)
	}

	csaManagers := sets.New(string(fieldOwner), crRegressionFieldManager, beforeFirstApplyFieldManager)

	patch, err := csaupgrade.UpgradeManagedFieldsPatch(obj, csaManagers, string(fieldOwner), opts...)
	if err != nil {
		return err
	}

	if patch != nil {
		return r.Patch(ctx, obj, client.RawPatch(types.JSONPatchType, patch))
	}

	// No work to be done - already upgraded
	return nil
}
