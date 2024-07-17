package stas

import (
	stasv1alpha1 "github.com/statnett/image-scanner-operator/api/stas/v1alpha1"
	stasv1alpha1ac "github.com/statnett/image-scanner-operator/internal/client/applyconfiguration/stas/v1alpha1"
)

func newContainerImageStatusPatch(cis *stasv1alpha1.ContainerImageScan) *stasv1alpha1ac.ContainerImageScanApplyConfiguration {
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

	return stasv1alpha1ac.ContainerImageScan(cis.Name, cis.Namespace).
		WithStatus(status)
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
