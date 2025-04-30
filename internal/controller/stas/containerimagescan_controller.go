package stas

import (
	"context"
	"fmt"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	metav1ac "k8s.io/client-go/applyconfigurations/meta/v1"
	kstatus "sigs.k8s.io/cli-utils/pkg/kstatus/status"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/source"

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
	EventChan chan event.GenericEvent
}

//+kubebuilder:rbac:groups=stas.statnett.no,resources=containerimagescans,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=stas.statnett.no,resources=containerimagescans/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=stas.statnett.no,resources=containerimagescans/finalizers,verbs=update
//+kubebuilder:rbac:groups=batch,resources=jobs,verbs=create;delete,namespace=image-scanner

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile
func (r *ContainerImageScanReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	fn := func(c context.Context) (ctrl.Result, error) {
		cis := &stasv1alpha1.ContainerImageScan{}
		if err := r.Get(ctx, req.NamespacedName, cis); err != nil {
			return ctrl.Result{}, staserrors.Ignore(err, apierrors.IsNotFound)
		}

		if r.ActiveScanJobLimit > 0 {
			count, err := r.activeScanJobCount(ctx)
			if err != nil {
				return ctrl.Result{}, err
			}

			if count >= r.ActiveScanJobLimit {
				// Max number of active scan jobs reached. Requeue request.
				return ctrl.Result{Requeue: true}, nil
			}
		}

		return r.reconcile(ctx, cis)
	}

	return controller.Reconcile(ctx, fn)
}

func (r *ContainerImageScanReconciler) activeScanJobCount(ctx context.Context) (int, error) {
	listOps := []client.ListOption{
		client.InNamespace(r.ScanJobNamespace),
		client.MatchingFields{indexJobCondition: jobNotFinished},
	}

	list := &batchv1.JobList{}
	if err := r.List(ctx, list, listOps...); err != nil {
		return 0, err
	}

	return len(list.Items), nil
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
				predicate.Not(cisScannedInInterval(r.ScanInterval)),
			)).
		WithEventFilter(predicate.And(predicates...)).
		WatchesRawSource(source.Channel(r.EventChan, &handler.EnqueueRequestForObject{})).
		Complete(r)
}

func (r *ContainerImageScanReconciler) reconcile(ctx context.Context, cis *stasv1alpha1.ContainerImageScan) (ctrl.Result, error) {
	logf.FromContext(ctx).Info("Reconciling")

	result := ctrl.Result{}

	scanJob, err := r.newScanJob(ctx, cis)
	if err != nil {
		return result, err
	}

	// Jobs are highly immutable, so not attempting to update
	err = r.Create(ctx, scanJob)
	if err != nil {
		if apierrors.IsAlreadyExists(err) {
			// Job already exists; delete it and requeue
			err = r.Delete(ctx, scanJob, client.PropagationPolicy(metav1.DeletePropagationBackground))
			result.Requeue = true
		}

		return result, err
	}

	return result, newContainerImageStatusPatch(cis).
		withCondition(
			metav1ac.Condition().
				WithType(string(kstatus.ConditionReconciling)).
				WithStatus(metav1.ConditionTrue).
				WithReason("ScanJobCreated").
				WithMessage(fmt.Sprintf("Job '%s' created to scan image.", scanJob.Name)),
		).
		apply(ctx, r.Client)
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
