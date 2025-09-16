package stas

import (
	"context"
	"fmt"
	"slices"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	metav1ac "k8s.io/client-go/applyconfigurations/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	stasv1alpha1 "github.com/statnett/image-scanner-operator/api/stas/v1alpha1"
	stasv1alpha1ac "github.com/statnett/image-scanner-operator/internal/client/applyconfiguration/stas/v1alpha1"
	"github.com/statnett/image-scanner-operator/internal/config"
	"github.com/statnett/image-scanner-operator/internal/config/feature"
)

func newContainerImageStatusPatch(cis *stasv1alpha1.ContainerImageScan) *containerImageScanStatusPatch {
	status := stasv1alpha1ac.ContainerImageScanStatus().
		WithObservedGeneration(cis.Generation).
		WithLastScanJobUID(cis.Status.LastScanJobUID)
	status.LastScanTime = cis.Status.LastScanTime
	status.LastSuccessfulScanTime = cis.Status.LastSuccessfulScanTime

	if cis.Status.VulnerabilitySummary != nil {
		status = status.WithVulnerabilitySummary(
			stasv1alpha1ac.VulnerabilitySummary().
				WithSeverityCount(cis.Status.VulnerabilitySummary.SeverityCount).
				WithFixedCount(cis.Status.VulnerabilitySummary.FixedCount).
				WithUnfixedCount(cis.Status.VulnerabilitySummary.UnfixedCount),
		)
	}

	if len(cis.Status.Vulnerabilities) > 0 {
		status.Vulnerabilities = make([]stasv1alpha1ac.VulnerabilityApplyConfiguration, len(cis.Status.Vulnerabilities))
		for i, v := range cis.Status.Vulnerabilities {
			status.Vulnerabilities[i] = *vulnerabilityPatch(v)
		}
	}

	return &containerImageScanStatusPatch{
		cis: cis,
		patch: stasv1alpha1ac.ContainerImageScan(cis.Name, cis.Namespace).
			WithStatus(status),
	}
}

type containerImageScanStatusPatch struct {
	cis         *stasv1alpha1.ContainerImageScan
	patch       *stasv1alpha1ac.ContainerImageScanApplyConfiguration
	minSeverity *stasv1alpha1.Severity
}

func (p *containerImageScanStatusPatch) withCondition(c *metav1ac.ConditionApplyConfiguration) *containerImageScanStatusPatch {
	p.patch.Status.
		WithConditions(NewConditionsPatch(p.cis.Status.Conditions, c)...)

	return p
}

func (p *containerImageScanStatusPatch) withScanJob(jobUID types.UID, successful bool, scanTime metav1.Time) *containerImageScanStatusPatch {
	p.patch.Status.
		WithLastScanJobUID(jobUID).
		WithLastScanTime(scanTime)

	if successful {
		p.patch.Status.
			WithLastSuccessfulScanTime(scanTime)
	}

	return p
}

func (p *containerImageScanStatusPatch) withResults(vulnerabilities []stasv1alpha1.Vulnerability, summary *stasv1alpha1.VulnerabilitySummary, minSeverity *stasv1alpha1.Severity) *containerImageScanStatusPatch {
	p.minSeverity = minSeverity

	if !config.DefaultFeatureGate.Enabled(feature.NoCISStatusVulnerabilities) {
		p.patch.Status.Vulnerabilities = make([]stasv1alpha1ac.VulnerabilityApplyConfiguration, len(vulnerabilities))
		for i, v := range vulnerabilities {
			p.patch.Status.Vulnerabilities[i] = *vulnerabilityPatch(v)
		}
	}

	p.patch.Status.
		WithVulnerabilitySummary(stasv1alpha1ac.VulnerabilitySummary().
			WithSeverityCount(summary.SeverityCount).
			WithFixedCount(summary.FixedCount).
			WithUnfixedCount(summary.UnfixedCount))

	return p
}

func (p *containerImageScanStatusPatch) apply(ctx context.Context, c client.Client) error {
	if err := upgradeStatusManagedFields(ctx, c, p.cis); err != nil {
		return fmt.Errorf("when upgrading status managed fields: %w", err)
	}

	if p.minSeverity == nil {
		if err := c.Status().Patch(ctx, p.cis, applyPatch{p.patch}, client.FieldValidation(metav1.FieldValidationStrict), client.ForceOwnership, fieldOwner); err != nil {
			return fmt.Errorf("when patching status: %w", err)
		}

		return nil
	}

	var err error
	// Repeat until resource fits in api-server by increasing minimum severity on failure.
	for severity := *p.minSeverity; severity <= stasv1alpha1.MaxSeverity; severity++ {
		p.patch.Status.Vulnerabilities = slices.DeleteFunc(p.patch.Status.Vulnerabilities, func(v stasv1alpha1ac.VulnerabilityApplyConfiguration) bool {
			return *v.Severity < severity
		})

		err = c.Status().Patch(ctx, p.cis, applyPatch{p.patch}, client.FieldValidation(metav1.FieldValidationStrict), client.ForceOwnership, fieldOwner)
		if !isResourceTooLargeError(err) {
			break
		}
	}

	if err != nil {
		return fmt.Errorf("when patching status: %w", err)
	}

	return nil
}

func vulnerabilityPatch(v stasv1alpha1.Vulnerability) *stasv1alpha1ac.VulnerabilityApplyConfiguration {
	return stasv1alpha1ac.Vulnerability().
		WithVulnerabilityID(v.VulnerabilityID).
		WithPkgName(v.PkgName).
		WithInstalledVersion(v.InstalledVersion).
		WithSeverity(v.Severity).
		WithPkgPath(v.PkgPath).
		WithFixedVersion(v.FixedVersion).
		WithTitle(v.Title).
		WithPrimaryURL(v.PrimaryURL)
}
