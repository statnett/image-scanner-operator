package stas

import (
	"context"
	"fmt"
	"maps"
	"slices"
	"strings"

	openreportsv1alpha1 "github.com/openreports/reports-api/apis/openreports.io/v1alpha1"
	openreportsv1alpha1ac "github.com/openreports/reports-api/pkg/client/applyconfiguration/openreports.io/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	stasv1alpha1 "github.com/statnett/image-scanner-operator/api/stas/v1alpha1"
)

func newPolicyReportPatch(cis *stasv1alpha1.ContainerImageScan) *policyReportPatch {
	return &policyReportPatch{
		cis: cis,
		patch: openreportsv1alpha1ac.Report(cis.Name, cis.Namespace).
			WithScope(
				corev1.ObjectReference{
					APIVersion: cis.Spec.Workload.APIVersion,
					Kind:       cis.Spec.Workload.Kind,
					Name:       cis.Spec.Workload.Name,
					UID:        cis.Spec.Workload.UID,
				},
			),
	}
}

type policyReportPatch struct {
	cis             *stasv1alpha1.ContainerImageScan
	patch           *openreportsv1alpha1ac.ReportApplyConfiguration
	vulnerabilities []stasv1alpha1.Vulnerability
	minSeverity     stasv1alpha1.Severity
}

func (p *policyReportPatch) withResults(vulnerabilities []stasv1alpha1.Vulnerability, summary *stasv1alpha1.VulnerabilitySummary, minSeverity *stasv1alpha1.Severity) *policyReportPatch {
	p.vulnerabilities = vulnerabilities
	p.minSeverity = ptr.Deref(minSeverity, stasv1alpha1.MinSeverity)

	p.patch.
		WithSummary(openreportsv1alpha1ac.ReportSummary().
			WithSkip(int(summary.SeverityCount[stasv1alpha1.SeverityUnknown.String()])).
			WithWarn(int(summary.SeverityCount[stasv1alpha1.SeverityLow.String()] + summary.SeverityCount[stasv1alpha1.SeverityMedium.String()])).
			WithFail(int(summary.SeverityCount[stasv1alpha1.SeverityHigh.String()] + summary.SeverityCount[stasv1alpha1.SeverityCritical.String()])))

	return p
}

func (p *policyReportPatch) apply(ctx context.Context, c client.Client, scheme *runtime.Scheme) error {
	if err := SetControllerReference(p.cis, p.patch.ObjectMetaApplyConfiguration, scheme); err != nil {
		return err
	}

	var err error
	// Repeat until resource fits in api-server by increasing minimum severity on failure.
	for severity := p.minSeverity; severity <= stasv1alpha1.MaxSeverity; severity++ {
		p.vulnerabilities = slices.DeleteFunc(p.vulnerabilities, func(v stasv1alpha1.Vulnerability) bool {
			return v.Severity < severity
		})

		p.patch.Results = make([]openreportsv1alpha1ac.ReportResultApplyConfiguration, len(p.vulnerabilities))
		for i, v := range p.vulnerabilities {
			p.patch.Results[i] = *policyReportResultPatch(v)
		}

		err = c.Apply(ctx, p.patch, client.ForceOwnership, fieldOwner)
		if !isResourceTooLargeError(err) {
			break
		}
	}

	if err != nil {
		return fmt.Errorf("when applying policy report: %w", err)
	}

	return nil
}

func policyReportResultPatch(v stasv1alpha1.Vulnerability) *openreportsv1alpha1ac.ReportResultApplyConfiguration {
	properties := map[string]string{
		"pkgName":          v.PkgName,
		"pkgPath":          v.PkgPath,
		"installedVersion": v.InstalledVersion,
		"fixedVersion":     v.FixedVersion,
		"primaryURL":       v.PrimaryURL,
	}

	// Remove properties with empty values to compact report
	maps.DeleteFunc(properties, func(k string, v string) bool {
		return len(v) == 0
	})

	report := openreportsv1alpha1ac.ReportResult().
		WithCategory("vulnerability scan").
		WithSource("image-scanner").
		WithPolicy(v.VulnerabilityID).
		WithResult(severityToPolicyResult(v.Severity)).
		WithDescription(v.Title).
		WithProperties(properties)

	if s, ok := severityToPolicyResultSeverity(v.Severity); ok {
		report = report.
			WithSeverity(s)
	}

	return report
}

func severityToPolicyResultSeverity(severity stasv1alpha1.Severity) (openreportsv1alpha1.ResultSeverity, bool) {
	switch severity {
	case stasv1alpha1.SeverityUnknown:
		return "", false
	default:
		return openreportsv1alpha1.ResultSeverity(strings.ToLower(severity.String())), true
	}
}

func severityToPolicyResult(severity stasv1alpha1.Severity) openreportsv1alpha1.Result {
	switch severity {
	case stasv1alpha1.SeverityUnknown:
		return "skip"
	case stasv1alpha1.SeverityLow, stasv1alpha1.SeverityMedium:
		return "warn"
	default:
		return "fail"
	}
}
