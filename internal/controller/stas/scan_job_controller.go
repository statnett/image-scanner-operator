package stas

import (
	"context"
	"errors"
	"fmt"
	"io"
	"slices"
	"strings"

	openreportsv1alpha1 "github.com/openreports/reports-api/apis/openreports.io/v1alpha1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	eventsv1 "k8s.io/api/events/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	metav1ac "k8s.io/client-go/applyconfigurations/meta/v1"
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
	"github.com/statnett/image-scanner-operator/internal/config/feature"
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
// Must add policyreports delete verb and containerimagescans/finalizers update verb to satisfy
// https://kubernetes.io/docs/reference/access-authn-authz/admission-controllers/#ownerreferencespermissionenforcement
//+kubebuilder:rbac:groups="openreports.io",resources=reports,verbs=get;list;watch;create;patch;delete
//+kubebuilder:rbac:groups=stas.statnett.no,resources=containerimagescans/finalizers,verbs=update

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
		Watches(&openreportsv1alpha1.Report{}, handler.Funcs{}). // Watches reports with empty handler to ensure informer creation for metrics collection
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
		return newContainerImageStatusPatch(cis).
			withCondition(
				metav1ac.Condition().
					WithType(string(kstatus.ConditionStalled)).
					WithStatus(metav1.ConditionTrue).
					WithReason(stasv1alpha1.ReasonScanReportDecodeError).
					WithMessage(fmt.Sprintf("error decoding scan report JSON from job '%s': %s", job.Name, err)),
			).
			withScanJob(job.UID, false, metav1.Now()).
			apply(ctx, r.Client)
	}

	slices.SortFunc(vulnerabilities, stasv1alpha1.BySeverity)

	minSeverity := stasv1alpha1.MinSeverity
	if cis.Spec.MinSeverity != nil {
		minSeverity = *cis.Spec.MinSeverity
	}

	summary := vulnerabilitySummary(vulnerabilities, minSeverity)

	if config.DefaultFeatureGate.Enabled(feature.PolicyReport) {
		err = newPolicyReportPatch(cis).
			withResults(vulnerabilities, summary, &minSeverity).
			apply(ctx, r.Client, r.Scheme)
		if err != nil {
			return err
		}
	}

	return newContainerImageStatusPatch(cis).
		withScanJob(job.UID, true, metav1.Now()).
		withResults(vulnerabilities, summary, &minSeverity).
		apply(ctx, r.Client)
}

func isResourceTooLargeError(err error) bool {
	return apierrors.IsRequestEntityTooLargeError(err) ||
		apierrors.IsInternalError(err) &&
			(strings.Contains(err.Error(), "ResourceExhausted") ||
				strings.Contains(err.Error(), "request is too large"))
}

func (r *ScanJobReconciler) reconcileFailedJob(ctx context.Context, job *batchv1.Job, log io.Reader, cis *stasv1alpha1.ContainerImageScan) error {
	logBytes, err := io.ReadAll(log)
	if err != nil {
		return err
	}

	return newContainerImageStatusPatch(cis).
		withCondition(
			metav1ac.Condition().
				WithType(string(kstatus.ConditionStalled)).
				WithStatus(metav1.ConditionTrue).
				WithReason("Error").
				WithMessage(string(logBytes)),
		).
		withScanJob(job.UID, false, metav1.Now()).
		apply(ctx, r.Client)
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

	switch {
	case isJobComplete(job):
		logf.FromContext(ctx).V(1).Info("Patching CIS status", "jobCondition", batchv1.JobComplete)
		return r.reconcileCompleteJob(ctx, job, logs, cis)
	case isJobFailed(job):
		logf.FromContext(ctx).V(1).Info("Patching CIS status", "jobCondition", batchv1.JobFailed)
		return r.reconcileFailedJob(ctx, job, logs, cis)
	default:
		return fmt.Errorf("don't know how to handle job status %q", job.Status.String())
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

func vulnerabilitySummary(vulnerabilities []stasv1alpha1.Vulnerability, minSeverity stasv1alpha1.Severity) *stasv1alpha1.VulnerabilitySummary {
	severityCount := make(map[string]int32)
	for severity := minSeverity; severity <= stasv1alpha1.MaxSeverity; severity++ {
		severityCount[severity.String()] = 0
	}

	var fixedCount, unfixedCount int32

	for _, vuln := range vulnerabilities {
		severityCount[vuln.Severity.String()] += 1

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
