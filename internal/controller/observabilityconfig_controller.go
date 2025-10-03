/*
Copyright 2025 shtsukada.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"
	"sort"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/tools/record"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	observabilityv1alpha1 "github.com/shtsukada/cloudnative-observability-operator/api/v1alpha1"
	conditions "github.com/shtsukada/cloudnative-observability-operator/internal/shared/conditions"
	tel "github.com/shtsukada/cloudnative-observability-operator/internal/shared/telemetry"
)

// ObservabilityConfigReconciler reconciles a ObservabilityConfig object
type ObservabilityConfigReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

// +kubebuilder:rbac:groups=observability.shtsukada.dev,resources=observabilityconfigs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=observability.shtsukada.dev,resources=observabilityconfigs/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=observability.shtsukada.dev,resources=observabilityconfigs/finalizers,verbs=update
// +kubebuilder:rbac:groups="",resources=services;configmaps;secrets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=deployments;daemonsets;statefulsets,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the ObservabilityConfig object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.21.0/pkg/reconcile
func (r *ObservabilityConfigReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	var oc observabilityv1alpha1.ObservabilityConfig
	if err := r.Get(ctx, req.NamespacedName, &oc); err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	orig := oc.DeepCopy()

	setProgressing(&oc, conditions.ReasonReconciling, "reconcile started")

	if oc.Spec.Endpoint == "" {
		conditions.Emit(r.Recorder, &oc, corev1.EventTypeWarning, conditions.ReasonErrInvalid, "spec.endpoint must not be empty")
		setDegraded(&oc, conditions.ReasonErrInvalid, "spec.endpoint must not be empty")
		oc.Status.ObservedGeneration = oc.Generation
		if err := r.Status().Patch(ctx, &oc, client.MergeFrom(orig)); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}

	changed, err := r.applyDesired(ctx, &oc)
	if err != nil {
		reason := conditions.ClassifyApplyError(err)
		conditions.Emit(r.Recorder, &oc, corev1.EventTypeWarning, reason, "apply failed:%v", err)
		setDegraded(&oc, reason, err.Error())
		oc.Status.ObservedGeneration = oc.Generation
		_ = r.Status().Patch(ctx, &oc, client.MergeFrom(orig))
		return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
	}
	if changed {
		conditions.Emit(r.Recorder, &oc, corev1.EventTypeNormal, conditions.ReasonApplySucceeded, "changes applied, waiting to settle")
		setProgressing(&oc, conditions.ReasonWaitingForDeployment, "changes applied, waiting to settle")
		oc.Status.ObservedGeneration = oc.Generation
		if err := r.Status().Patch(ctx, &oc, client.MergeFrom(orig)); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{RequeueAfter: 2 * time.Second}, nil
	}
	conditions.Emit(r.Recorder, &oc, corev1.EventTypeNormal, conditions.ReasonDeploymentAvailable, "resources are in sync")
	setReady(&oc, conditions.ReasonDeploymentAvailable, "resources are in sync")

	if oc.Status.ObservedGeneration != oc.Generation {
		oc.Status.ObservedGeneration = oc.Generation
	}

	if err := r.Status().Patch(ctx, &oc, client.MergeFrom(orig)); err != nil {
		return ctrl.Result{}, err
	}

	log.V(1).Info("reconciled ObservabilityConfig", "name", req.NamespacedName)
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ObservabilityConfigReconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.Recorder = mgr.GetEventRecorderFor("cloudnative-observability-operator")
	wrapped := tel.WrapReconciler("ObservabilityConfig", r)
	return ctrl.NewControllerManagedBy(mgr).
		For(&observabilityv1alpha1.ObservabilityConfig{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.ConfigMap{}).
		Named("observabilityconfig").
		Complete(wrapped)
}

func setCondition(oc *observabilityv1alpha1.ObservabilityConfig, t string, status metav1.ConditionStatus, reason, msg string) {
	cond := metav1.Condition{
		Type:               t,
		Status:             status,
		Reason:             reason,
		Message:            msg,
		LastTransitionTime: metav1.NewTime(time.Now()),
		ObservedGeneration: oc.GetGeneration(),
	}
	apimeta.SetStatusCondition(&oc.Status.Conditions, cond)
}

func setReady(oc *observabilityv1alpha1.ObservabilityConfig, reason, msg string) {
	setCondition(oc, "Ready", metav1.ConditionTrue, reason, msg)
	oc.Status.Phase = "Ready"
	oc.Status.Reason = reason
}

func setProgressing(oc *observabilityv1alpha1.ObservabilityConfig, reason, msg string) {
	setCondition(oc, "Ready", metav1.ConditionFalse, reason, msg)
	oc.Status.Phase = "Reconciling"
	oc.Status.Reason = reason
}

func setDegraded(oc *observabilityv1alpha1.ObservabilityConfig, reason, msg string) {
	setCondition(oc, "Ready", metav1.ConditionFalse, reason, msg)
	oc.Status.Phase = "Error"
	oc.Status.Reason = reason
}

func (r *ObservabilityConfigReconciler) desiredDeployment(oc *observabilityv1alpha1.ObservabilityConfig) *appsv1.Deployment {
	labels := map[string]string{
		"app.kubernetes.io/name":       "oc-sidecar",
		"app.kubernetes.io/managed-by": "cloudnative-observability-operator",
		"obscfg":                       oc.Name,
	}

	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      oc.Name + "-oc-sidecar",
			Namespace: oc.Namespace,
			Labels:    labels,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: ptr.To(int32(1)),
			Selector: &metav1.LabelSelector{MatchLabels: labels},
			Strategy: appsv1.DeploymentStrategy{
				Type: appsv1.RollingUpdateDeploymentStrategyType,
				RollingUpdate: &appsv1.RollingUpdateDeployment{
					MaxUnavailable: ptr.To(intstr.FromString("25%")),
					MaxSurge:       ptr.To(intstr.FromString("25%")),
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: labels},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:            "oc-agent",
							Image:           "example/oc-agent:dummy",
							ImagePullPolicy: corev1.PullIfNotPresent,
							Ports: []corev1.ContainerPort{
								{Name: "otlp-grpc", ContainerPort: 4317},
								{Name: "otlp-http", ContainerPort: 4318},
							},
							Env: []corev1.EnvVar{
								{Name: "OTEL_EXPORTER_OTLP_ENDPOINT", Value: oc.Spec.Endpoint},
							},
						},
					},
				},
			},
		},
	}
	return dep
}

func (r *ObservabilityConfigReconciler) desiredService(oc *observabilityv1alpha1.ObservabilityConfig) *corev1.Service {
	labels := map[string]string{
		"app.kubernetes.io/name": "oc-sidecar",
		"obscfg":                 oc.Name,
	}
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      oc.Name + "-oc",
			Namespace: oc.Namespace,
			Labels:    labels,
		},
		Spec: corev1.ServiceSpec{
			Selector: labels,
			Ports: []corev1.ServicePort{
				{Name: "otlp-grpc", Port: 4317, TargetPort: intstr.FromInt(4317)},
				{Name: "otlp-http", Port: 4318, TargetPort: intstr.FromInt(4318)},
			},
			Type: corev1.ServiceTypeClusterIP,
		},
	}
}

func (r *ObservabilityConfigReconciler) applyDesired(ctx context.Context, oc *observabilityv1alpha1.ObservabilityConfig) (bool, error) {
	wantDep := r.desiredDeployment(oc)
	haveDep := &appsv1.Deployment{ObjectMeta: wantDep.ObjectMeta}

	op, err := controllerutil.CreateOrUpdate(ctx, r.Client, haveDep, func() error {
		if err := controllerutil.SetControllerReference(oc, haveDep, r.Scheme); err != nil {
			return err
		}
		r.mutateDeployment(haveDep, wantDep)
		return nil
	})
	if err != nil {
		return false, err
	}

	wantSvc := r.desiredService(oc)
	haveSvc := &corev1.Service{ObjectMeta: wantSvc.ObjectMeta}

	op2, err := controllerutil.CreateOrUpdate(ctx, r.Client, haveSvc, func() error {
		if err := controllerutil.SetControllerReference(oc, haveSvc, r.Scheme); err != nil {
			return err
		}

		clusterIP := haveSvc.Spec.ClusterIP
		clusterIPs := haveSvc.Spec.ClusterIPs
		r.mutateService(haveSvc, wantSvc)
		haveSvc.Spec.ClusterIP = clusterIP
		if len(clusterIPs) > 0 {
			haveSvc.Spec.ClusterIPs = clusterIPs
		}
		return nil
	})
	if err != nil {
		return false, err
	}

	changed := (op != controllerutil.OperationResultNone || op2 != controllerutil.OperationResultNone)
	return changed, nil
}

func (r *ObservabilityConfigReconciler) mutateDeployment(have, want *appsv1.Deployment) {
	have.Labels = want.Labels
	have.Spec.Selector = want.Spec.Selector
	have.Spec.Replicas = want.Spec.Replicas
	have.Spec.Strategy = want.Spec.Strategy

	r.normPodSpec(&want.Spec.Template.Spec)
	r.normPodSpec(&have.Spec.Template.Spec)

	have.Spec.Template = want.Spec.Template
}

func (r *ObservabilityConfigReconciler) mutateService(have, want *corev1.Service) {
	have.Labels = want.Labels
	have.Spec.Selector = want.Spec.Selector

	sort.Slice(want.Spec.Ports, func(i, j int) bool { return want.Spec.Ports[i].Name < want.Spec.Ports[j].Name })
	sort.Slice(have.Spec.Ports, func(i, j int) bool { return have.Spec.Ports[i].Name < have.Spec.Ports[j].Name })
	have.Spec.Ports = want.Spec.Ports

	have.Spec.Type = want.Spec.Type
}

func (r *ObservabilityConfigReconciler) normPodSpec(p *corev1.PodSpec) {
	for i := range p.Containers {
		if p.Containers[i].ImagePullPolicy == "" {
			p.Containers[i].ImagePullPolicy = corev1.PullIfNotPresent
		}
		sort.Slice(p.Containers[i].Env, func(a, b int) bool { return p.Containers[i].Env[a].Name < p.Containers[i].Env[b].Name })
		sort.Slice(p.Containers[i].Ports, func(a, b int) bool { return p.Containers[i].Ports[a].Name < p.Containers[i].Ports[b].Name })
	}
}
