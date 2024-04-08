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
	eventsv1 "k8s.io/api/events/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	kstatus "sigs.k8s.io/cli-utils/pkg/kstatus/status"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/json"

	stasv1alpha1 "github.com/statnett/image-scanner-operator/api/stas/v1alpha1"
	"github.com/statnett/image-scanner-operator/internal/config"
	"github.com/statnett/image-scanner-operator/internal/controller"
	staserrors "github.com/statnett/image-scanner-operator/internal/errors"
	"github.com/statnett/image-scanner-operator/internal/pod"
	"github.com/statnett/image-scanner-operator/internal/trivy"
)

var backoffContainerStateReasons = map[string]struct{}{
	"ImagePullBackOff": {},
	"ErrImagePull":     {},
}

// ScanJobReconciler reconciles a finished image scan Job object.
type ScanJobReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	config.Config
	pod.LogsReader
}

//+kubebuilder:rbac:groups=batch,resources=jobs,verbs=get;list;watch,namespace=image-scanner
//+kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch,namespace=image-scanner
//+kubebuilder:rbac:groups="",resources=pods/log,verbs=get;list,namespace=image-scanner
//+kubebuilder:rbac:groups="events.k8s.io",resources=events,verbs=get;list;watch

// SetupWithManager sets up the controller with the Manager.
func (r *ScanJobReconciler) SetupWithManager(mgr ctrl.Manager) error {
	err := ctrl.NewControllerManagedBy(mgr).
		For(&batchv1.Job{},
			builder.WithPredicates(
				managedByImageScanner,
				inNamespacePredicate(r.ScanJobNamespace),
				jobIsFinished,
				ignoreDeletionPredicate(),
			)).
		Complete(r.reconcile())
	if err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		Named("backOffScanJobPod").
		WithEventFilter(inNamespacePredicate(r.ScanJobNamespace)).
		Watches(
			&eventsv1.Event{},
			handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, obj client.Object) []reconcile.Request {
				e := obj.(*eventsv1.Event)
				return []reconcile.Request{
					{NamespacedName: types.NamespacedName{
						Name:      e.Regarding.Name,
						Namespace: e.Regarding.Namespace,
					}},
				}
			}),
			builder.WithPredicates(
				eventRegardingKind("Pod"),
				eventReason("BackOff"),
			),
		).
		Complete(r.reconcileBackOffJobPod())
}

func (r *ScanJobReconciler) reconcileBackOffJobPod() reconcile.Func {
	return func(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
		logf.FromContext(ctx).Info("Reconciling")

		fn := func(c context.Context) (ctrl.Result, error) {
			p := &corev1.Pod{}
			if err := r.Get(ctx, req.NamespacedName, p); err != nil {
				return ctrl.Result{}, staserrors.Ignore(err, apierrors.IsNotFound)
			}

			var stateWaiting *corev1.ContainerStateWaiting

			for _, cs := range p.Status.ContainerStatuses {
				if csw := cs.State.Waiting; csw != nil {
					if _, ok := backoffContainerStateReasons[csw.Reason]; ok {
						stateWaiting = csw
						break
					}
				}
			}

			if stateWaiting == nil {
				expectedReasons := make([]string, 0, len(backoffContainerStateReasons))
				for k := range backoffContainerStateReasons {
					expectedReasons = append(expectedReasons, k)
				}

				logf.FromContext(ctx).V(1).Info("no waiting state found", "expectedReasons", expectedReasons)
				// Pod (in controller cache) has not yet reached waiting state. Requeue event
				return ctrl.Result{Requeue: true}, nil
			}

			podController := metav1.GetControllerOf(p)
			if podController == nil {
				return ctrl.Result{}, fmt.Errorf("no controller found for pod %q", p.Name)
			}

			if podController.Kind != "Job" {
				return ctrl.Result{}, nil
			}

			job := &batchv1.Job{}
			if err := r.Get(ctx, client.ObjectKey{Namespace: req.Namespace, Name: podController.Name}, job); err != nil {
				return ctrl.Result{}, staserrors.Ignore(err, apierrors.IsNotFound)
			}

			return ctrl.Result{}, r.reconcileBackOffJob(ctx, job, stateWaiting.Message)
		}

		return controller.Reconcile(ctx, fn)
	}
}

func (r *ScanJobReconciler) reconcileBackOffJob(ctx context.Context, job *batchv1.Job, errMsg string) error {
	if err := r.Delete(ctx, job, client.PropagationPolicy(metav1.DeletePropagationBackground)); err != nil {
		return err
	}

	cis, err := r.getScanJobCIS(ctx, job)
	if err != nil {
		return staserrors.Ignore(err, staserrors.IsNotFound)
	}

	return r.reconcileFailedJob(ctx, job, strings.NewReader(errMsg), cis)
}

func (r *ScanJobReconciler) reconcileCompleteJob(ctx context.Context, job *batchv1.Job, log io.Reader, cis *stasv1alpha1.ContainerImageScan) error {
	var vulnerabilities []stasv1alpha1.Vulnerability

	err := json.NewDecoderCaseSensitivePreserveInts(log).Decode(&vulnerabilities)
	if err != nil {
		cleanCis := cis.DeepCopy()

		condition := metav1.Condition{
			Type:    string(kstatus.ConditionStalled),
			Status:  metav1.ConditionTrue,
			Reason:  stasv1alpha1.ReasonScanReportDecodeError,
			Message: fmt.Sprintf("error decoding scan report JSON from job '%s': %s", job.Name, err),
		}
		meta.SetStatusCondition(&cis.Status.Conditions, condition)
		meta.RemoveStatusCondition(&cis.Status.Conditions, string(kstatus.ConditionReconciling))
		cis.Status.LastScanTime = ptr.To(metav1.Now())
		cis.Status.LastScanJobUID = job.UID

		err = r.Status().Patch(ctx, cis, client.MergeFrom(cleanCis))
		if err != nil {
			logf.FromContext(ctx).Error(err, "when patching status", "condition", condition)
		}

		return err
	}

	sort.Sort(stasv1alpha1.BySeverity(vulnerabilities))

	minSeverity := stasv1alpha1.MinSeverity
	if cis.Spec.MinSeverity != nil {
		minSeverity, err = stasv1alpha1.NewSeverity(*cis.Spec.MinSeverity)
		if err != nil {
			return err
		}
	}

	return r.updateCISStatus(ctx, job, cis, vulnerabilities, minSeverity)
}

func (r *ScanJobReconciler) updateCISStatus(ctx context.Context, job *batchv1.Job, cis *stasv1alpha1.ContainerImageScan, vulnerabilities []stasv1alpha1.Vulnerability, minSeverity stasv1alpha1.Severity) error {
	cleanCis := cis.DeepCopy()
	now := metav1.Now()

	cis.Status.VulnerabilitySummary = vulnerabilitySummary(vulnerabilities, minSeverity)
	// Clear any conditions since we now have a successful scan report
	cis.Status.Conditions = nil
	cis.Status.LastScanTime = &now
	cis.Status.LastScanJobUID = job.UID
	cis.Status.LastSuccessfulScanTime = &now

	// Repeat until resource fits in api-server by increasing minimum severity on failure.
	for severity := minSeverity; severity <= stasv1alpha1.MaxSeverity; severity++ {
		var err error

		cis.Status.Vulnerabilities, err = filterVulnerabilities(vulnerabilities, severity)
		if err != nil {
			return err
		}

		err = r.Status().Patch(ctx, cis, client.MergeFrom(cleanCis))
		if err == nil || !isResourceTooLargeError(err) {
			return err
		}
	}

	return nil
}

func isResourceTooLargeError(err error) bool {
	return apierrors.IsInternalError(err) &&
		(strings.Contains(err.Error(), "ResourceExhausted") ||
			strings.Contains(err.Error(), "request is too large"))
}

func (r *ScanJobReconciler) reconcileFailedJob(ctx context.Context, job *batchv1.Job, log io.Reader, cis *stasv1alpha1.ContainerImageScan) error {
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
	cis.Status.LastScanJobUID = job.UID

	err = r.Status().Patch(ctx, cis, client.MergeFrom(cleanCis))
	if err != nil {
		logf.FromContext(ctx).Error(err, "when patching status", "condition", condition)
	}

	return err
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
	logf.FromContext(ctx).V(1).Info("Reconciling", "status", job.Status)

	cis, err := r.getScanJobCIS(ctx, job)
	if err != nil {
		return staserrors.Ignore(err, staserrors.IsNotFound)
	}

	if job.UID == cis.Status.LastScanJobUID {
		// We already reconciled this job; no point doing it again
		return nil
	}

	logs, err := r.getScanJobLogs(ctx, job)
	if err != nil {
		switch {
		case staserrors.IsNotFound(err), staserrors.IsScanJobContainerWaiting(err):
			logf.FromContext(ctx).V(1).Info("Error while fetching logs", "error", err)
			return r.reconcileFailedJob(ctx, job, strings.NewReader(err.Error()), cis)
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

	switch jc := jobCondition(job); jc {
	case batchv1.JobComplete:
		logf.FromContext(ctx).V(1).Info("Patching CIS status", "jobCondition", jc)
		return r.reconcileCompleteJob(ctx, job, logs, cis)
	case batchv1.JobFailed:
		logf.FromContext(ctx).V(1).Info("Patching CIS status", "jobCondition", jc)
		return r.reconcileFailedJob(ctx, job, logs, cis)
	default:
		return fmt.Errorf("I don't know how to handle job status %q", jc)
	}
}

func (r *ScanJobReconciler) getScanJobCIS(ctx context.Context, job *batchv1.Job) (*stasv1alpha1.ContainerImageScan, error) {
	cisList := &stasv1alpha1.ContainerImageScanList{}

	listOps := []client.ListOption{
		client.InNamespace(job.Labels[stasv1alpha1.LabelStatnettControllerNamespace]),
		client.MatchingFields{indexUID: job.Labels[stasv1alpha1.LabelStatnettControllerUID]},
	}
	if err := r.List(ctx, cisList, listOps...); err != nil {
		return nil, err
	}

	switch len(cisList.Items) {
	case 0:
		return nil, staserrors.NewNotFound(fmt.Sprintf("no CISes found for job %q", job.Name))
	case 1:
	default:
		return nil, errors.New("expected number of container image scans to be 1")
	}

	return &cisList.Items[0], nil
}

func (r *ScanJobReconciler) getScanJobLogs(ctx context.Context, job *batchv1.Job) (io.ReadCloser, error) {
	// Find Job pod
	selector, err := metav1.LabelSelectorAsSelector(job.Spec.Selector)
	if err != nil {
		return nil, err
	}

	podList := &corev1.PodList{}
	if err = r.List(ctx, podList, client.MatchingLabelsSelector{Selector: selector}); err != nil {
		return nil, err
	}

	var pods []corev1.Pod

	for _, p := range podList.Items {
		if p.Status.Reason != "Evicted" {
			pods = append(pods, p)
		}
	}

	switch len(pods) {
	case 0:
		return nil, staserrors.NewNotFound(fmt.Sprintf("no pods found for job %q", job.Name))
	case 1:
	default:
		return nil, fmt.Errorf("expected number of job pods to be 1, got %d ", len(pods))
	}

	jobPod := pods[0]

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

func filterVulnerabilities(orig []stasv1alpha1.Vulnerability, minSeverity stasv1alpha1.Severity) ([]stasv1alpha1.Vulnerability, error) {
	var filtered []stasv1alpha1.Vulnerability

	for _, v := range orig {
		severity, err := stasv1alpha1.NewSeverity(v.Severity)
		if err != nil {
			return nil, err
		}

		if severity >= minSeverity {
			filtered = append(filtered, v)
		}
	}

	return filtered, nil
}

func vulnerabilitySummary(vulnerabilities []stasv1alpha1.Vulnerability, minSeverity stasv1alpha1.Severity) *stasv1alpha1.VulnerabilitySummary {
	severityCount := make(map[string]int32)
	for severity := minSeverity; severity <= stasv1alpha1.MaxSeverity; severity++ {
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
