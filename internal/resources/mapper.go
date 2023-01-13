package resources

import (
	"fmt"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// NotNamespacedError is returned if the mapper can find the resource, but it is not namespaced.
type NotNamespacedError struct {
	GroupResource schema.GroupResource
}

func (e *NotNamespacedError) Error() string {
	return fmt.Sprintf("resource %q is not namespaced", e.GroupResource)
}

// ResourceKindMapper allows callers to map resources to kinds using apiserver metadata at runtime
type ResourceKindMapper struct {
	RestMapper meta.RESTMapper
}

// NamespacedKindFor maps namespaced GR to GVK by using a DynamicRESTMapper
// to discover resource types at runtime. Will return an error if the resource isn't namespaced.
func (m *ResourceKindMapper) NamespacedKindFor(gr schema.GroupResource) (schema.GroupVersionKind, error) {
	resource := gr.WithVersion("")
	kinds, err := m.RestMapper.KindsFor(resource)
	if err != nil {
		return schema.GroupVersionKind{}, err
	}
	// Not sure if this can ever happen? RestMapper.KindsFor does not specify if no
	// matches will error out or not, so just making sure we do not get issues.
	if len(kinds) == 0 {
		return schema.GroupVersionKind{}, &meta.NoResourceMatchError{PartialResource: resource}
	}

	// KindsFor returns the list of potential kinds in priority order
	gvk := kinds[0]
	mapping, err := m.RestMapper.RESTMapping(gvk.GroupKind())
	if err != nil {
		return schema.GroupVersionKind{}, fmt.Errorf("while attempting to determine if kind is namespaced: %w", err)
	}
	if mapping.Scope.Name() != meta.RESTScopeNameNamespace {
		return schema.GroupVersionKind{}, &NotNamespacedError{GroupResource: gr}
	}

	return gvk, nil
}

// NamespacedKindsForResources maps namespaced GRs to GVKs.
// The format used for GR is "resource.group", i.e. "replicasets.apps"
func (m *ResourceKindMapper) NamespacedKindsForResources(resources ...string) ([]schema.GroupVersionKind, error) {
	var kinds []schema.GroupVersionKind
	for _, r := range resources {
		gr := schema.ParseGroupResource(r)
		kind, err := m.NamespacedKindFor(gr)
		if err != nil {
			return nil, err
		}
		kinds = append(kinds, kind)
	}
	return kinds, nil
}
