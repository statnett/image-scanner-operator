package stas

import (
	"context"
	"fmt"
	"time"

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
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	stasv1alpha1 "github.com/statnett/image-scanner-operator/api/stas/v1alpha1"
	"github.com/statnett/image-scanner-operator/internal/config"
	"github.com/statnett/image-scanner-operator/internal/controller"
	staserrors "github.com/statnett/image-scanner-operator/internal/errors"
	"github.com/statnett/image-scanner-operator/internal/trivy"
)

// ContainerImageScanReconciler reconciles a ContainerImageScan object.
type ContainerImageScanReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	config.Config
}

//+kubebuilder:rbac:groups=stas.statnett.no,resources=containerimagescans,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=stas.statnett.no,resources=containerimagescans/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=stas.statnett.no,resources=containerimagescans/finalizers,verbs=update
//+kubebuilder:rbac:groups=batch,resources=jobs,verbs=create

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile
func (r *ContainerImageScanReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logf.FromContext(ctx).Info("Reconciling")

	fn := func(c context.Context) (ctrl.Result, error) {
		cis := &stasv1alpha1.ContainerImageScan{}
		if err := r.Get(ctx, req.NamespacedName, cis); err != nil {
			return ctrl.Result{}, staserrors.Ignore(err, apierrors.IsNotFound)
		}

		var timeUntilNextScan time.Duration

		switch {
		case cis.Status.ObservedGeneration != cis.Generation:
			// Spec has changed; must scan
		case cis.Status.LastScanTime.IsZero():
			// New resource; must scan
		default:
			d := time.Until(cis.Status.LastScanTime.Add(r.ScanInterval))
			if d > 0 {
				timeUntilNextScan = d
			}
		}

		var err error
		if timeUntilNextScan == 0 {
			err = r.reconcile(ctx, cis)
			timeUntilNextScan = r.ScanInterval
		}

		return ctrl.Result{RequeueAfter: timeUntilNextScan}, err
	}

	return controller.Reconcile(ctx, fn)
}

// SetupWithManager sets up the controller with the Manager.
func (r *ContainerImageScanReconciler) SetupWithManager(mgr ctrl.Manager) error {
	var predicates []predicate.Predicate
	if r.ScanNamespaceExcludeRegexp != nil {
		predicates = append(predicates, predicate.Not(namespaceMatchRegexp(r.ScanNamespaceExcludeRegexp)))
	}

	if r.ScanNamespaceIncludeRegexp != nil {
		predicates = append(predicates, namespaceMatchRegexp(r.ScanNamespaceIncludeRegexp))
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&stasv1alpha1.ContainerImageScan{},
			builder.WithPredicates(
				predicate.GenerationChangedPredicate{},
				ignoreDeletionPredicate(),
			)).
		WithEventFilter(predicate.And(predicates...)).
		Complete(r)
}

func (r *ContainerImageScanReconciler) reconcile(ctx context.Context, cis *stasv1alpha1.ContainerImageScan) error {
	cleanCis := cis.DeepCopy()

	scanJob, err := r.newScanJob(ctx, cis)
	if err != nil {
		return err
	}

	// Jobs are highly immutable, so not attempting to update
	err = r.Create(ctx, scanJob)
	if err != nil {
		if apierrors.IsAlreadyExists(err) {
			// Don't create duplicate jobs
			return nil
		}

		return err
	}

	condition := metav1.Condition{
		Type:    string(kstatus.ConditionReconciling),
		Status:  metav1.ConditionTrue,
		Reason:  "ScanJobCreated",
		Message: fmt.Sprintf("Job '%s' created to scan image.", scanJob.Name),
	}
	meta.SetStatusCondition(&cis.Status.Conditions, condition)
	meta.RemoveStatusCondition(&cis.Status.Conditions, string(kstatus.ConditionStalled))

	cis.Status.ObservedGeneration = cis.Generation

	return r.Status().Patch(ctx, cis, client.MergeFrom(cleanCis))
}

func (r *ContainerImageScanReconciler) newScanJob(ctx context.Context, cis *stasv1alpha1.ContainerImageScan) (*batchv1.Job, error) {
	var nodeNames []string

	for _, or := range cis.OwnerReferences {
		pod := &corev1.Pod{}
		if err := r.Get(ctx, client.ObjectKey{Name: or.Name, Namespace: cis.Namespace}, pod); err != nil {
			if apierrors.IsNotFound(err) {
				// Owner might have been deleted; continue to next owner
				continue
			}

			return nil, err
		}

		if pod.Spec.NodeName != "" {
			nodeNames = append(nodeNames, pod.Spec.NodeName)
		}
	}

	job, err := trivy.NewImageScanJob(r.Config).
		OnPreferredNodes(nodeNames...).
		ForCIS(cis)
	if err != nil {
		return nil, err
	}

	return job, nil
}
