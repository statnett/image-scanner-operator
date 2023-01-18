package stas

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sort"
	"strings"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	kstatus "sigs.k8s.io/cli-utils/pkg/kstatus/status"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/json"

	stasv1alpha1 "github.com/statnett/image-scanner-operator/api/stas/v1alpha1"
	"github.com/statnett/image-scanner-operator/internal/controller"
	staserrors "github.com/statnett/image-scanner-operator/internal/errors"
	"github.com/statnett/image-scanner-operator/internal/pod"
	"github.com/statnett/image-scanner-operator/internal/trivy"
	"github.com/statnett/image-scanner-operator/pkg/operator"
)

// ScanJobReconciler reconciles a finished image scan Job object.
type ScanJobReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	operator.Config
	pod.LogsReader
}

//+kubebuilder:rbac:groups=batch,resources=jobs,verbs=get;list;watch
//+kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch
//+kubebuilder:rbac:groups="",resources=pods/log,verbs=get;list

// SetupWithManager sets up the controller with the Manager.
func (r *ScanJobReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&batchv1.Job{},
			builder.WithPredicates(
				managedByImageScanner,
				inNamespacePredicate(r.ScanJobNamespace),
				jobIsFinished,
				ignoreDeletionPredicate(),
			)).
		Complete(r.reconcile())
}

func (r *ScanJobReconciler) reconcileCompleteJob(ctx context.Context, jobName string, log io.Reader, cis *stasv1alpha1.ContainerImageScan) error {
	var (
		cleanCis        = cis.DeepCopy()
		vulnerabilities []stasv1alpha1.Vulnerability
		now             = metav1.Now()
	)

	err := json.NewDecoderCaseSensitivePreserveInts(log).Decode(&vulnerabilities)
	if err != nil {
		cis = cleanCis.DeepCopy()

		condition := metav1.Condition{
			Type:    string(kstatus.ConditionStalled),
			Status:  metav1.ConditionTrue,
			Reason:  stasv1alpha1.ReasonScanReportDecodeError,
			Message: fmt.Sprintf("error decoding scan report JSON from job '%s': %s", jobName, err),
		}
		meta.SetStatusCondition(&cis.Status.Conditions, condition)
		meta.RemoveStatusCondition(&cis.Status.Conditions, string(kstatus.ConditionReconciling))
		cis.Status.LastScanTime = &now
		cis.Status.LastScanJobName = jobName
		err = r.Status().Patch(ctx, cis, client.MergeFrom(cleanCis))
		return err
	}

	sort.Sort(stasv1alpha1.BySeverity(vulnerabilities))
	cis.Status.Vulnerabilities = vulnerabilities

	minSeverity := stasv1alpha1.SeverityUnknown
	if cis.Spec.MinSeverity != nil {
		minSeverity, err = stasv1alpha1.NewSeverity(*cis.Spec.MinSeverity)
		if err != nil {
			return err
		}
	}

	cis.Status.VulnerabilitySummary = vulnerabilitySummary(vulnerabilities, minSeverity)
	// Clear any conditions since we now have a successful scan report
	cis.Status.Conditions = nil
	cis.Status.LastScanTime = &now
	cis.Status.LastScanJobName = jobName
	cis.Status.LastSuccessfulScanTime = &now

	err = r.Status().Patch(ctx, cis, client.MergeFrom(cleanCis))
	if err != nil && isResourceTooLargeError(err) {
		cis = cleanCis.DeepCopy()

		condition := metav1.Condition{
			Type:    string(kstatus.ConditionStalled),
			Status:  metav1.ConditionTrue,
			Reason:  stasv1alpha1.ReasonVulnerabilityOverflow,
			Message: fmt.Sprintf("vulnerability report is too large to fit in API: %s", err),
		}
		meta.SetStatusCondition(&cis.Status.Conditions, condition)
		meta.RemoveStatusCondition(&cis.Status.Conditions, string(kstatus.ConditionReconciling))
		cis.Status.LastScanTime = &now
		cis.Status.LastScanJobName = jobName
		err = r.Status().Patch(ctx, cis, client.MergeFrom(cleanCis))
	}

	return err
}

func isResourceTooLargeError(err error) bool {
	return apierrors.IsInternalError(err) &&
		(strings.Contains(err.Error(), "ResourceExhausted") ||
			strings.Contains(err.Error(), "request is too large"))
}

func (r *ScanJobReconciler) reconcileFailedJob(ctx context.Context, jobName string, log io.Reader, cis *stasv1alpha1.ContainerImageScan) error {
	cleanCis := cis.DeepCopy()

	logBytes, err := io.ReadAll(log)
	if err != nil {
		return err
	}

	condition := metav1.Condition{
		Type:    string(kstatus.ConditionStalled),
		Status:  metav1.ConditionTrue,
		Reason:  "Error",
		Message: string(logBytes),
	}
	meta.SetStatusCondition(&cis.Status.Conditions, condition)
	meta.RemoveStatusCondition(&cis.Status.Conditions, string(kstatus.ConditionReconciling))

	now := metav1.Now()
	cis.Status.LastScanTime = &now
	cis.Status.LastScanJobName = jobName

	return r.Status().Patch(ctx, cis, client.MergeFrom(cleanCis))
}

func (r *ScanJobReconciler) reconcile() reconcile.Func {
	return func(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
		logf.FromContext(ctx).Info("Reconciling")

		fn := func(c context.Context) (ctrl.Result, error) {
			job := &batchv1.Job{}
			if err := r.Get(ctx, req.NamespacedName, job); err != nil {
				return ctrl.Result{}, staserrors.Ignore(err, apierrors.IsNotFound)
			}

			return ctrl.Result{}, r.reconcileJob(ctx, job)
		}

		return controller.Reconcile(ctx, fn)
	}
}

func (r *ScanJobReconciler) reconcileJob(ctx context.Context, job *batchv1.Job) error {
	cisList := &stasv1alpha1.ContainerImageScanList{}

	listOps := []client.ListOption{
		client.InNamespace(job.Labels[stasv1alpha1.LabelStatnettControllerNamespace]),
		client.MatchingFields{indexUID: job.Labels[stasv1alpha1.LabelStatnettControllerUID]},
	}
	if err := r.List(ctx, cisList, listOps...); err != nil {
		return err
	}

	switch len(cisList.Items) {
	case 0:
		// CIS deleted; nothing more to do
		return nil
	case 1:
	default:
		return errors.New("expected number of container image scans to be 1")
	}

	cis := &cisList.Items[0]

	if job.Name == cis.Status.LastScanJobName {
		// We already reconciled this job; no point doing it again
		return nil
	}

	logs, err := r.getScanJobLogs(ctx, job)
	if err != nil {
		switch {
		case staserrors.IsJobPodNotFound(err), staserrors.IsScanJobContainerWaiting(err):
			return r.reconcileFailedJob(ctx, job.Name, strings.NewReader(err.Error()), cis)
		default:
			return err
		}
	}

	defer func(podLogs io.ReadCloser) {
		err := podLogs.Close()
		if err != nil {
			logf.FromContext(ctx).Error(err, "could not close log stream")
		}
	}(logs)

	if job.Status.Succeeded > 0 {
		return r.reconcileCompleteJob(ctx, job.Name, logs, cis)
	} else {
		return r.reconcileFailedJob(ctx, job.Name, logs, cis)
	}
}

func (r *ScanJobReconciler) getScanJobLogs(ctx context.Context, job *batchv1.Job) (io.ReadCloser, error) {
	// Find Job pod
	selector, err := metav1.LabelSelectorAsSelector(job.Spec.Selector)
	if err != nil {
		return nil, err
	}

	pods := &corev1.PodList{}
	if err = r.List(ctx, pods, client.MatchingLabelsSelector{Selector: selector}); err != nil {
		return nil, err
	}

	switch len(pods.Items) {
	case 0:
		return nil, staserrors.NewJobPodNotFound(job.Name)
	case 1:
	default:
		return nil, fmt.Errorf("expected number of job pods to be 1, got %d ", len(pods.Items))
	}

	jobPod := pods.Items[0]

	var scanJobContainerStatus corev1.ContainerStatus

	for _, cs := range jobPod.Status.ContainerStatuses {
		if cs.Name == trivy.ScanJobContainerName {
			scanJobContainerStatus = cs
			break
		}
	}

	if scanJobContainerStatus.State.Waiting != nil {
		return nil, staserrors.NewScanJobContainerWaiting(*scanJobContainerStatus.State.Waiting)
	}
	// Get logs from Job pod
	return r.GetLogs(ctx, client.ObjectKeyFromObject(&jobPod), trivy.ScanJobContainerName)
}

func vulnerabilitySummary(vulnerabilities []stasv1alpha1.Vulnerability, minSeverity stasv1alpha1.Severity) *stasv1alpha1.VulnerabilitySummary {
	severityCount := make(map[string]int32)
	for severity := minSeverity; severity <= stasv1alpha1.SeverityCritical; severity++ {
		severityCount[severity.String()] = 0
	}

	var fixedCount, unfixedCount int32

	for _, vuln := range vulnerabilities {
		severityCount[vuln.Severity] += 1

		if vuln.FixedVersion != "" {
			fixedCount++
		} else {
			unfixedCount++
		}
	}

	return &stasv1alpha1.VulnerabilitySummary{
		SeverityCount: severityCount,
		FixedCount:    fixedCount,
		UnfixedCount:  unfixedCount,
	}
}
