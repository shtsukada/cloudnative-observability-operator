package controller

import (
	"context"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	apiv1alpha1 "github.com/shtsukada/cloudnative-observability-operator/api/v1alpha1"
	conditions "github.com/shtsukada/cloudnative-observability-operator/internal/shared/conditions"
)

const finalizerName = "grpcburner.finalizers.observability.shtsukada.dev"

// +kubebuilder:rbac:groups=observability.shtsukada.dev,resources=grpcburners,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=observability.shtsukada.dev,resources=grpcburners/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=observability.shtsukada.dev,resources=grpcburners/finalizers,verbs=update
// +kubebuilder:rbac:groups="",resources=serviceaccounts;services;events,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete

type GrpcBurnerReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

func (r *GrpcBurnerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	var gb apiv1alpha1.GrpcBurner
	if err := r.Get(ctx, req.NamespacedName, &gb); err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	if !gb.ObjectMeta.DeletionTimestamp.IsZero() {
		if controllerutil.ContainsFinalizer(&gb, finalizerName) {
			controllerutil.RemoveFinalizer(&gb, finalizerName)
			if err := r.Update(ctx, &gb); err != nil {
				return ctrl.Result{}, err
			}
			r.Recorder.Event(&gb, corev1.EventTypeNormal, "Finalized", "Finalizer cleanup completed")
		}
		return ctrl.Result{}, nil
	}

	if !controllerutil.ContainsFinalizer(&gb, finalizerName) {
		controllerutil.AddFinalizer(&gb, finalizerName)
		if err := r.Update(ctx, &gb); err != nil {
			return ctrl.Result{}, err
		}
	}

	r.setCondition(&gb, apiv1alpha1.ConditionProgressing, metav1.ConditionTrue, conditions.ReasonReconciling, "Reconciling desired state")
	if err := r.Status().Update(ctx, &gb); err != nil {
		logger.V(1).Info("status update (progressing) failed", "err", err)
	}

	sa := desiredServiceAccount(&gb)
	svc := desiredService(&gb)
	deploy := desiredDeployment(&gb)

	if err := r.createOrUpdate(ctx, &gb, sa, func() error { return nil }); err != nil {
		return r.fail(&gb, err)
	}
	if err := r.createOrUpdate(ctx, &gb, svc, func() error { return nil }); err != nil {
		return r.fail(&gb, err)
	}
	if err := r.createOrUpdate(ctx, &gb, deploy, func() error { return nil }); err != nil {
		return r.fail(&gb, err)
	}

	var d appsv1.Deployment
	if err := r.Get(ctx, types.NamespacedName{Name: deploy.Name, Namespace: deploy.Namespace}, &d); err == nil {
		gb.Status.ReadyReplicas = d.Status.ReadyReplicas
		_ = r.Status().Update(ctx, &gb)
		if d.Status.ReadyReplicas == ptr.Deref(deploy.Spec.Replicas, 1) {
			conditions.Emit(r.Recorder, &gb, corev1.EventTypeNormal, conditions.ReasonDeploymentAvailable, "deployment available: ready=%d", d.Status.ReadyReplicas)
			gb.SetCondition(apiv1alpha1.ConditionReady, metav1.ConditionTrue, conditions.ReasonDeploymentAvailable, "Deployment ready")
			gb.SetCondition(apiv1alpha1.ConditionProgressing, metav1.ConditionFalse, "Stable", "Reconcile stable")
			_ = r.Status().Update(ctx, &gb)
		} else {
			conditions.Emit(r.Recorder, &gb, corev1.EventTypeNormal, conditions.ReasonDeploymentUnavailable, "deployment progressing: ready=%d/%d", d.Status.ReadyReplicas, ptr.Deref(deploy.Spec.Replicas, 1))
		}
	}
	return ctrl.Result{}, nil
}

func (r *GrpcBurnerReconciler) createOrUpdate(ctx context.Context, owner *apiv1alpha1.GrpcBurner, obj client.Object, mutate func() error) error {
	if err := controllerutil.SetControllerReference(owner, obj, r.Scheme); err != nil {
		return err
	}
	key := types.NamespacedName{Name: obj.GetName(), Namespace: obj.GetNamespace()}
	current := obj.DeepCopyObject().(client.Object)

	if err := r.Get(ctx, key, current); err != nil {
		if apierrors.IsNotFound(err) {
			if err := mutate(); err != nil {
				return err
			}
			if err := r.Create(ctx, obj); err != nil {
				return err
			}
			r.Recorder.Event(owner, corev1.EventTypeNormal, "Created", fmt.Sprintf("%T %q created", obj, key.Name))
			return nil
		}
		return err
	}
	if err := mutate(); err != nil {
		return err
	}
	obj.SetResourceVersion(current.GetResourceVersion())
	if err := r.Update(ctx, obj); err != nil {
		return err
	}
	r.Recorder.Event(owner, corev1.EventTypeNormal, "Updated", fmt.Sprintf("%T %q updated", obj, key.Name))
	return nil
}

func (r *GrpcBurnerReconciler) setCondition(gb *apiv1alpha1.GrpcBurner, cond string, status metav1.ConditionStatus, reason, msg string) {
	gb.SetCondition(cond, status, reason, msg)
}

func (r *GrpcBurnerReconciler) fail(gb *apiv1alpha1.GrpcBurner, err error) (ctrl.Result, error) {
	reason := conditions.ClassifyApplyError(err)
	conditions.Emit(r.Recorder, gb, corev1.EventTypeWarning, reason, "apply failed:%v", err)
	gb.SetCondition(apiv1alpha1.ConditionDegraded, metav1.ConditionTrue, reason, err.Error())
	gb.SetCondition(apiv1alpha1.ConditionReady, metav1.ConditionFalse, reason, "Not ready")
	_ = r.Status().Update(context.Background(), gb)
	return ctrl.Result{}, err
}

func (r *GrpcBurnerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.Recorder = mgr.GetEventRecorderFor("cloudnative-observability-operator")
	return ctrl.NewControllerManagedBy(mgr).
		For(&apiv1alpha1.GrpcBurner{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Service{}).
		Owns(&corev1.ServiceAccount{}).
		Complete(r)
}

func mergeMap(dst, src map[string]string) map[string]string {
	if dst == nil && src == nil {
		return nil
	}
	if dst == nil {
		dst = map[string]string{}
	}
	for k, v := range src {
		dst[k] = v
	}
	return dst
}
