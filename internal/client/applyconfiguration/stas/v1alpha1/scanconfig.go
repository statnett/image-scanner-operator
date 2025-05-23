// Code generated by controller-gen. DO NOT EDIT.

package v1alpha1

import (
	stasv1alpha1 "github.com/statnett/image-scanner-operator/api/stas/v1alpha1"
)

// ScanConfigApplyConfiguration represents a declarative configuration of the ScanConfig type for use
// with apply.
type ScanConfigApplyConfiguration struct {
	MinSeverity   *stasv1alpha1.Severity `json:"minSeverity,omitempty"`
	IgnoreUnfixed *bool                  `json:"ignoreUnfixed,omitempty"`
}

// ScanConfigApplyConfiguration constructs a declarative configuration of the ScanConfig type for use with
// apply.
func ScanConfig() *ScanConfigApplyConfiguration {
	return &ScanConfigApplyConfiguration{}
}

// WithMinSeverity sets the MinSeverity field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the MinSeverity field is set to the value of the last call.
func (b *ScanConfigApplyConfiguration) WithMinSeverity(value stasv1alpha1.Severity) *ScanConfigApplyConfiguration {
	b.MinSeverity = &value
	return b
}

// WithIgnoreUnfixed sets the IgnoreUnfixed field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the IgnoreUnfixed field is set to the value of the last call.
func (b *ScanConfigApplyConfiguration) WithIgnoreUnfixed(value bool) *ScanConfigApplyConfiguration {
	b.IgnoreUnfixed = &value
	return b
}
