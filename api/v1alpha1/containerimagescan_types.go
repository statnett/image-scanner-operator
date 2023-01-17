package v1alpha1

import (
	"github.com/distribution/distribution/reference"
	"github.com/opencontainers/go-digest"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kstatus "sigs.k8s.io/cli-utils/pkg/kstatus/status"
)

const (
	ReasonVulnerabilityOverflow        = "VulnerabilityOverflow"
	WorkloadAnnotationKeyIgnoreUnfixed = "image-scanner.statnett.no/ignore-unfixed"
)

type Image struct {
	Name   string        `json:"name"`
	Digest digest.Digest `json:"digest"`
}

type Workload struct {
	metav1.GroupKind `json:",inline"`
	Name             string `json:"name"`
	ContainerName    string `json:"containerName"`
}

type ScanConfig struct {
	// MinSeverity sets the minimum vulnerability severity included when scanning the image.
	//+kubebuilder:validation:Enum={UNKNOWN,LOW,MEDIUM,HIGH,CRITICAL}
	MinSeverity *string `json:"minSeverity,omitempty"`
	// IgnoreUnfixed set to true will report only fixed vulnerabilities when scanning the image.
	IgnoreUnfixed *bool `json:"ignoreUnfixed,omitempty"`
}

type VulnerabilitySummary struct {
	// VulnerabilitySummary is a summary of vulnerability counts grouped by Severity.
	// +mapType=atomic
	SeverityCount map[string]int32 `json:"severityCount,omitempty"`
	// FixedCount is the total number of fixed vulnerabilities where a patch is available.
	FixedCount int32 `json:"fixedCount"`
	// UnfixedCount is the total number of vulnerabilities where no patch is yet available.
	UnfixedCount int32 `json:"unfixedCount"`
}

func (vs *VulnerabilitySummary) GetSeverityCount() map[string]int32 {
	if vs == nil {
		return nil
	}

	return vs.SeverityCount
}

func (in *Image) Canonical() (reference.Canonical, error) {
	named, err := reference.ParseNamed(in.Name)
	if err != nil {
		return nil, err
	}

	return reference.WithDigest(named, in.Digest)
}

func (cis ContainerImageScan) HasVulnerabilityOverflow() bool {
	if cis.Status.ObservedGeneration != cis.Generation {
		// CIS is still under reconciliation
		return false
	}

	stalledCondition := meta.FindStatusCondition(cis.Status.Conditions, string(kstatus.ConditionStalled))
	if stalledCondition == nil {
		return false
	}

	return stalledCondition.Reason == ReasonVulnerabilityOverflow
}

// ImageScanSpec represents the specification for the container image scan.
type ImageScanSpec struct {
	Image      `json:",inline"`
	ScanConfig `json:",inline"`
}

// ContainerImageScanSpec contains a resolved container image in use by owning workload.
type ContainerImageScanSpec struct {
	ImageScanSpec `json:",inline"`
	Tag           string   `json:"tag,omitempty"`
	Workload      Workload `json:"workload"`
}

// ContainerImageScanStatus defines the observed state of ContainerImageScan.
type ContainerImageScanStatus struct {
	// ObservedGeneration is the generation observed by the image scanner operator.
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
	// LastScanTime is the timestamp for the last attempt to scan the image.
	LastScanTime *metav1.Time `json:"lastScanTime,omitempty"`
	// LastScanJobName is the name of the scan job that last (successfully) updated the status.
	LastScanJobName string `json:"lastScanJobName,omitempty"`
	// LastSuccessfulScanTime is the timestamp for the last successful scan of the image.
	LastSuccessfulScanTime *metav1.Time `json:"lastSuccessfulScanTime,omitempty"`
	// Conditions represent the latest available observations of an object's state.
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// Vulnerabilities contains the image scan result.
	// NOTE: This is currently in an experimental state, and is subject to breaking changes.
	// +listType=atomic
	Vulnerabilities []Vulnerability `json:"vulnerabilities,omitempty"`

	// VulnerabilitySummary is a summary of detected vulnerabilities.
	VulnerabilitySummary *VulnerabilitySummary `json:"vulnerabilitySummary,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:shortName={cis}
//+kubebuilder:printcolumn:name="Image",type=string,JSONPath=`.spec.name`
//+kubebuilder:printcolumn:name="Digest",type=string,JSONPath=`.spec.digest`
//+kubebuilder:printcolumn:name="Tag",type=string,JSONPath=`.spec.tag`

// ContainerImageScan is the Schema for the containerImageScans API.
type ContainerImageScan struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ContainerImageScanSpec   `json:"spec,omitempty"`
	Status ContainerImageScanStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ContainerImageScanList contains a list of ContainerImageScan.
type ContainerImageScanList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ContainerImageScan `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ContainerImageScan{}, &ContainerImageScanList{})
}
