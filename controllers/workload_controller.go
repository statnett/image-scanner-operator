package controllers

import (
	"context"
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/pointer"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	stasv1alpha1 "github.com/statnett/image-scanner-operator/api/v1alpha1"
	"github.com/statnett/image-scanner-operator/internal/controller"
	staserrors "github.com/statnett/image-scanner-operator/internal/errors"
	"github.com/statnett/image-scanner-operator/internal/hash"
	"github.com/statnett/image-scanner-operator/pkg/operator"
)

const (
	ImageShortSHALength     = 5
	KubernetesNameMaxLength = 253
)

var noEventsEventHandler = handler.EnqueueRequestsFromMapFunc(func(o client.Object) []reconcile.Request {
	return nil
})

type PodReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	operator.Config
	WorkloadKinds []schema.GroupVersionKind
}

//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch
//+kubebuilder:rbac:groups=apps,resources=replicasets,verbs=get;list;watch
//+kubebuilder:rbac:groups=apps,resources=statefulsets,verbs=get;list;watch
//+kubebuilder:rbac:groups=apps,resources=daemonsets,verbs=get;list;watch
//+kubebuilder:rbac:groups=batch,resources=cronjobs,verbs=get;list;watch
//+kubebuilder:rbac:groups=batch,resources=jobs,verbs=get;list;watch
//+kubebuilder:rbac:groups="",resources=replicationcontrollers,verbs=get;list;watch
//+kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch

// SetupWithManager sets up the controllers with the Manager.
func (r *PodReconciler) SetupWithManager(mgr ctrl.Manager) error {
	groupKinds := make([]schema.GroupKind, len(r.WorkloadKinds))
	for i, k := range r.WorkloadKinds {
		groupKinds[i] = k.GroupKind()
	}

	bldr := ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Pod{},
			builder.WithPredicates(
				podContainerStatusImagesChanged(),
				predicate.Or(controllerInKinds(groupKinds...), noController),
				ignoreDeletionPredicate(),
			)).
		WithEventFilter(predicate.And(
			predicate.Not(systemNamespace),
			predicate.Not(managedByImageScanner),
		)).
		Watches(&source.Kind{Type: &stasv1alpha1.ContainerImageScan{}},
			&handler.EnqueueRequestForOwner{OwnerType: &corev1.Pod{}},
			builder.WithPredicates(
				predicate.Or(predicate.GenerationChangedPredicate{}, cisVulnerabilityOverflow),
				ignoreCreationPredicate(),
			))
	for _, kind := range r.WorkloadKinds {
		obj := &metav1.PartialObjectMetadata{}
		obj.SetGroupVersionKind(kind)
		bldr.Watches(&source.Kind{Type: obj}, noEventsEventHandler, builder.OnlyMetadata)
	}
	return bldr.Complete(r.reconcilePod())
}

func (r *PodReconciler) reconcilePod() reconcile.Func {
	return func(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
		fn := func(ctx context.Context) (ctrl.Result, error) {
			pod := &corev1.Pod{}
			if err := r.Get(ctx, req.NamespacedName, pod); err != nil {
				return ctrl.Result{}, staserrors.Ignore(err, apierrors.IsNotFound)
			}
			if pod.GetDeletionTimestamp() != nil {
				// Workload is about to be deleted. No need to do any Reconcile.
				return ctrl.Result{}, nil
			}
			err := r.reconcile(ctx, pod)
			return ctrl.Result{}, err
		}
		return controller.Reconcile(ctx, fn)
	}
}

func (r *PodReconciler) reconcile(ctx context.Context, pod *corev1.Pod) error {
	logf.FromContext(ctx).Info("Reconciling")

	podController, err := r.getControllerWorkloadOrSelf(ctx, pod)
	if err != nil {
		return err
	}

	images, err := containerImages(pod)
	if err != nil {
		return err
	}
	for containerName, image := range images {
		cis := &stasv1alpha1.ContainerImageScan{}
		cis.Namespace = pod.Namespace
		cis.Name = imageScanName(podController, containerName, image)
		mutateFn := func() error {
			cis.Labels = pod.GetLabels()

			cis.Spec.Workload.Group = podController.GetObjectKind().GroupVersionKind().Group
			cis.Spec.Workload.Kind = podController.GetObjectKind().GroupVersionKind().Kind
			cis.Spec.Workload.Name = podController.GetName()
			cis.Spec.Workload.ContainerName = containerName

			cis.Spec.Image = image
			var ignoreUnfixed *bool
			if v := podController.GetAnnotations()[stasv1alpha1.WorkloadAnnotationKeyIgnoreUnfixed]; v == "true" {
				ignoreUnfixed = pointer.Bool(true)
			}
			cis.Spec.IgnoreUnfixed = ignoreUnfixed

			if cis.HasVulnerabilityOverflow() {
				minSeverity := stasv1alpha1.MinSeverity
				if cis.Spec.MinSeverity != nil {
					minSeverity, err = stasv1alpha1.NewSeverity(*cis.Spec.MinSeverity)
					if err != nil {
						return err
					}
				}
				if minSeverity < stasv1alpha1.MaxSeverity {
					minSeverity++
					cis.Spec.MinSeverity = pointer.String(minSeverity.String())
				}
			}
			return controllerutil.SetOwnerReference(pod, cis, r.Scheme)
		}
		_, err = controllerutil.CreateOrUpdate(ctx, r.Client, cis, mutateFn)
		if err != nil {
			return err
		}

		err = r.garbageCollectObsoleteImageScans(ctx, pod, cis)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *PodReconciler) garbageCollectObsoleteImageScans(ctx context.Context, pod *corev1.Pod, wantCIS *stasv1alpha1.ContainerImageScan) error {
	CISes, err := r.getImageScansOwnedByPodContainer(ctx, pod, wantCIS.Spec.Workload.ContainerName)
	if err != nil {
		return err
	}
	for _, cis := range CISes {
		if cis.Name != wantCIS.Name {
			if err := r.Delete(ctx, &cis); err != nil {
				return err
			}
		}
	}
	return nil
}

func (r *PodReconciler) getImageScansOwnedByPodContainer(ctx context.Context, pod *corev1.Pod, containerName string) ([]stasv1alpha1.ContainerImageScan, error) {
	listOps := []client.ListOption{
		client.InNamespace(pod.Namespace),
		client.MatchingFields{indexOwnerUID: string(pod.UID)},
	}
	list := &stasv1alpha1.ContainerImageScanList{}
	if err := r.List(ctx, list, listOps...); err != nil {
		return nil, err
	}
	var CISes []stasv1alpha1.ContainerImageScan
	for _, cis := range list.Items {
		if cis.Spec.Workload.ContainerName == containerName {
			CISes = append(CISes, cis)
		}
	}
	return CISes, nil
}

func containerImages(pod *corev1.Pod) (map[string]stasv1alpha1.Image, error) {
	images := make(map[string]stasv1alpha1.Image)
	for _, containerStatus := range pod.Status.ContainerStatuses {
		if containerStatus.Image != "" && containerStatus.ImageID != "" {
			image, err := NewImageFromContainerStatus(containerStatus)
			if err != nil {
				return nil, err
			}
			images[containerStatus.Name] = image
		}
	}
	return images, nil
}

func imageScanName(controller client.Object, containerName string, image stasv1alpha1.Image) string {
	kindPart := strings.ToLower(controller.GetObjectKind().GroupVersionKind().Kind)
	imagePart := hash.NewString(image.Name, image.Digest)[0:ImageShortSHALength]
	nameFn := func(controllerName string) string {
		return fmt.Sprintf("%s-%s-%s-%s", kindPart, controllerName, containerName, imagePart)
	}

	controllerName := controller.GetName()
	name := nameFn(controllerName)
	if len(name) > KubernetesNameMaxLength {
		shortenControllerName := controllerName[0 : len(controllerName)-(len(name)-KubernetesNameMaxLength)]
		name = nameFn(shortenControllerName)
	}
	return name
}

func (r *PodReconciler) getControllerWorkloadOrSelf(ctx context.Context, controllee client.Object) (client.Object, error) {
	ref := metav1.GetControllerOf(controllee)
	if ref == nil {
		// No controller; just return the object
		return controllee, nil
	}

	refInWorkloadKinds, err := r.inWorkloadKinds(ref)
	if err != nil {
		return nil, err
	}
	if !refInWorkloadKinds {
		// Controller not in workload kinds; return the object
		return controllee, nil
	}

	c := &metav1.PartialObjectMetadata{}
	c.APIVersion = ref.APIVersion
	c.Kind = ref.Kind
	name := types.NamespacedName{
		Namespace: controllee.GetNamespace(),
		Name:      ref.Name,
	}
	if err = r.Get(ctx, name, c); err != nil {
		return nil, err
	}

	return r.getControllerWorkloadOrSelf(ctx, c)
}

func (r *PodReconciler) inWorkloadKinds(ref *metav1.OwnerReference) (bool, error) {
	gv, err := schema.ParseGroupVersion(ref.APIVersion)
	if err != nil {
		return false, err
	}
	gk := schema.GroupKind{Group: gv.Group, Kind: ref.Kind}

	for _, kind := range r.WorkloadKinds {
		if kind.Group == gk.Group && kind.Kind == gk.Kind {
			return true, nil
		}
	}
	return false, nil
}
