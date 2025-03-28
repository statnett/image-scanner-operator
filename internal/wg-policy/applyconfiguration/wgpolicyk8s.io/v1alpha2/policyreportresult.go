// Code generated by applyconfiguration-gen. DO NOT EDIT.

package v1alpha2

import (
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	metav1 "k8s.io/client-go/applyconfigurations/meta/v1"
	wgpolicyk8siov1alpha2 "sigs.k8s.io/wg-policy-prototypes/policy-report/apis/wgpolicyk8s.io/v1alpha2"
)

// PolicyReportResultApplyConfiguration represents a declarative configuration of the PolicyReportResult type for use
// with apply.
type PolicyReportResultApplyConfiguration struct {
	Source          *string                                     `json:"source,omitempty"`
	Policy          *string                                     `json:"policy,omitempty"`
	Rule            *string                                     `json:"rule,omitempty"`
	Category        *string                                     `json:"category,omitempty"`
	Severity        *wgpolicyk8siov1alpha2.PolicyResultSeverity `json:"severity,omitempty"`
	Timestamp       *v1.Timestamp                               `json:"timestamp,omitempty"`
	Result          *wgpolicyk8siov1alpha2.PolicyResult         `json:"result,omitempty"`
	Scored          *bool                                       `json:"scored,omitempty"`
	Subjects        []corev1.ObjectReference                    `json:"resources,omitempty"`
	SubjectSelector *metav1.LabelSelectorApplyConfiguration     `json:"resourceSelector,omitempty"`
	Description     *string                                     `json:"message,omitempty"`
	Properties      map[string]string                           `json:"properties,omitempty"`
}

// PolicyReportResultApplyConfiguration constructs a declarative configuration of the PolicyReportResult type for use with
// apply.
func PolicyReportResult() *PolicyReportResultApplyConfiguration {
	return &PolicyReportResultApplyConfiguration{}
}

// WithSource sets the Source field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Source field is set to the value of the last call.
func (b *PolicyReportResultApplyConfiguration) WithSource(value string) *PolicyReportResultApplyConfiguration {
	b.Source = &value
	return b
}

// WithPolicy sets the Policy field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Policy field is set to the value of the last call.
func (b *PolicyReportResultApplyConfiguration) WithPolicy(value string) *PolicyReportResultApplyConfiguration {
	b.Policy = &value
	return b
}

// WithRule sets the Rule field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Rule field is set to the value of the last call.
func (b *PolicyReportResultApplyConfiguration) WithRule(value string) *PolicyReportResultApplyConfiguration {
	b.Rule = &value
	return b
}

// WithCategory sets the Category field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Category field is set to the value of the last call.
func (b *PolicyReportResultApplyConfiguration) WithCategory(value string) *PolicyReportResultApplyConfiguration {
	b.Category = &value
	return b
}

// WithSeverity sets the Severity field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Severity field is set to the value of the last call.
func (b *PolicyReportResultApplyConfiguration) WithSeverity(value wgpolicyk8siov1alpha2.PolicyResultSeverity) *PolicyReportResultApplyConfiguration {
	b.Severity = &value
	return b
}

// WithTimestamp sets the Timestamp field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Timestamp field is set to the value of the last call.
func (b *PolicyReportResultApplyConfiguration) WithTimestamp(value v1.Timestamp) *PolicyReportResultApplyConfiguration {
	b.Timestamp = &value
	return b
}

// WithResult sets the Result field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Result field is set to the value of the last call.
func (b *PolicyReportResultApplyConfiguration) WithResult(value wgpolicyk8siov1alpha2.PolicyResult) *PolicyReportResultApplyConfiguration {
	b.Result = &value
	return b
}

// WithScored sets the Scored field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Scored field is set to the value of the last call.
func (b *PolicyReportResultApplyConfiguration) WithScored(value bool) *PolicyReportResultApplyConfiguration {
	b.Scored = &value
	return b
}

// WithSubjects adds the given value to the Subjects field in the declarative configuration
// and returns the receiver, so that objects can be build by chaining "With" function invocations.
// If called multiple times, values provided by each call will be appended to the Subjects field.
func (b *PolicyReportResultApplyConfiguration) WithSubjects(values ...corev1.ObjectReference) *PolicyReportResultApplyConfiguration {
	for i := range values {
		b.Subjects = append(b.Subjects, values[i])
	}
	return b
}

// WithSubjectSelector sets the SubjectSelector field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the SubjectSelector field is set to the value of the last call.
func (b *PolicyReportResultApplyConfiguration) WithSubjectSelector(value *metav1.LabelSelectorApplyConfiguration) *PolicyReportResultApplyConfiguration {
	b.SubjectSelector = value
	return b
}

// WithDescription sets the Description field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Description field is set to the value of the last call.
func (b *PolicyReportResultApplyConfiguration) WithDescription(value string) *PolicyReportResultApplyConfiguration {
	b.Description = &value
	return b
}

// WithProperties puts the entries into the Properties field in the declarative configuration
// and returns the receiver, so that objects can be build by chaining "With" function invocations.
// If called multiple times, the entries provided by each call will be put on the Properties field,
// overwriting an existing map entries in Properties field with the same key.
func (b *PolicyReportResultApplyConfiguration) WithProperties(entries map[string]string) *PolicyReportResultApplyConfiguration {
	if b.Properties == nil && len(entries) > 0 {
		b.Properties = make(map[string]string, len(entries))
	}
	for k, v := range entries {
		b.Properties[k] = v
	}
	return b
}
