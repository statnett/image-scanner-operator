package stas

import (
	"context"
	"fmt"
	"maps"
	"slices"
	"strings"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	policyv1alpha2 "sigs.k8s.io/wg-policy-prototypes/policy-report/apis/wgpolicyk8s.io/v1alpha2"

	stasv1alpha1 "github.com/statnett/image-scanner-operator/api/stas/v1alpha1"
	policyv1alpha2ac "github.com/statnett/image-scanner-operator/internal/wg-policy/applyconfiguration/wgpolicyk8s.io/v1alpha2"
)

func newPolicyReportPatch(cis *stasv1alpha1.ContainerImageScan) *policyReportPatch {
	return &policyReportPatch{
		cis: cis,
		patch: policyv1alpha2ac.PolicyReport(cis.Name, cis.Namespace).
			WithScope(
				corev1.ObjectReference{
					Kind: cis.Spec.Workload.Kind,
					Name: cis.Spec.Workload.Name,
				},
			),
	}
}

type policyReportPatch struct {
	cis             *stasv1alpha1.ContainerImageScan
	patch           *policyv1alpha2ac.PolicyReportApplyConfiguration
	vulnerabilities []stasv1alpha1.Vulnerability
	minSeverity     stasv1alpha1.Severity
}

func (p *policyReportPatch) withResults(vulnerabilities []stasv1alpha1.Vulnerability, summary *stasv1alpha1.VulnerabilitySummary, minSeverity stasv1alpha1.Severity) *policyReportPatch {
	p.vulnerabilities = vulnerabilities
	p.minSeverity = minSeverity

	p.patch.
		WithSummary(policyv1alpha2ac.PolicyReportSummary().
			WithSkip(int(summary.SeverityCount[stasv1alpha1.SeverityUnknown.String()])).
			WithWarn(int(summary.SeverityCount[stasv1alpha1.SeverityLow.String()] + summary.SeverityCount[stasv1alpha1.SeverityMedium.String()])).
			WithFail(int(summary.SeverityCount[stasv1alpha1.SeverityHigh.String()] + summary.SeverityCount[stasv1alpha1.SeverityCritical.String()])))

	return p
}

func (p *policyReportPatch) apply(ctx context.Context, c client.Client, scheme *runtime.Scheme) error {
	if err := SetControllerReference(p.cis, p.patch.ObjectMetaApplyConfiguration, scheme); err != nil {
		return err
	}

	report := &policyv1alpha2.PolicyReport{}
	report.Name = *p.patch.Name
	report.Namespace = *p.patch.Namespace

	var err error
	// Repeat until resource fits in api-server by increasing minimum severity on failure.
	for severity := p.minSeverity; severity <= stasv1alpha1.MaxSeverity; severity++ {
		p.vulnerabilities = slices.DeleteFunc(p.vulnerabilities, func(v stasv1alpha1.Vulnerability) bool {
			return v.Severity < severity
		})

		p.patch.Results = make([]policyv1alpha2ac.PolicyReportResultApplyConfiguration, len(p.vulnerabilities))
		for i, v := range p.vulnerabilities {
			p.patch.Results[i] = *policyReportResultPatch(v)
		}

		err = c.Patch(ctx, report, applyPatch{p.patch}, FieldValidationStrict, client.ForceOwnership, fieldOwner)
		if !isResourceTooLargeError(err) {
			break
		}
	}

	if err != nil {
		return fmt.Errorf("when applying policy report: %w", err)
	}

	return nil
}

func policyReportResultPatch(v stasv1alpha1.Vulnerability) *policyv1alpha2ac.PolicyReportResultApplyConfiguration {
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

	report := policyv1alpha2ac.PolicyReportResult().
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

func severityToPolicyResultSeverity(severity stasv1alpha1.Severity) (policyv1alpha2.PolicyResultSeverity, bool) {
	switch severity {
	case stasv1alpha1.SeverityUnknown:
		return "", false
	default:
		return policyv1alpha2.PolicyResultSeverity(strings.ToLower(severity.String())), true
	}
}

func severityToPolicyResult(severity stasv1alpha1.Severity) policyv1alpha2.PolicyResult {
	switch severity {
	case stasv1alpha1.SeverityUnknown:
		return "skip"
	case stasv1alpha1.SeverityLow, stasv1alpha1.SeverityMedium:
		return "warn"
	default:
		return "fail"
	}
}
