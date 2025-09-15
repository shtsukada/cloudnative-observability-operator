package controller

import (
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/ptr"

	apiv1alpha1 "github.com/shtsukada/cloudnative-observability-operator/api/v1alpha1"
)

func labels(gb *apiv1alpha1.GrpcBurner) map[string]string {
	return map[string]string{
		"app.kubernetes.io/name":       "grpcburner",
		"app.kubernetes.io/instance":   gb.Name,
		"app.kubernetes.io/managed-by": "cloudnative-observability-operator",
	}
}

func desiredServiceAccount(gb *apiv1alpha1.GrpcBurner) *corev1.ServiceAccount {
	return &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-sa", gb.Name),
			Namespace: gb.Namespace,
			Labels:    labels(gb),
		},
	}
}

func desiredService(gb *apiv1alpha1.GrpcBurner) *corev1.Service {
	lbl := labels(gb)

	ports := func() []corev1.ServicePort {
		if len(gb.Spec.Ports) == 0 {
			return []corev1.ServicePort{{
				Name:       "grpc",
				Port:       50051,
				Protocol:   corev1.ProtocolTCP,
				TargetPort: intstr.FromInt(50051),
			}}
		}
		out := make([]corev1.ServicePort, 0, len(gb.Spec.Ports))
		for _, p := range gb.Spec.Ports {
			svcPort := p.ContainerPort
			if p.ServicePort != nil {
				svcPort = *p.ServicePort
			}
			out = append(out, corev1.ServicePort{
				Name:       p.Name,
				Port:       svcPort,
				Protocol:   p.Protocol,
				TargetPort: intstr.FromInt(int(p.ContainerPort)),
			})
		}
		return out
	}()

	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-svc", gb.Name),
			Namespace: gb.Namespace,
			Labels:    lbl,
		},
		Spec: corev1.ServiceSpec{
			Selector: lbl,
			Ports:    ports,
		},
	}
}

func desiredDeployment(gb *apiv1alpha1.GrpcBurner) *appsv1.Deployment {
	lbl := labels(gb)

	replicas := int32(1)
	if gb.Spec.Replicas != nil {
		replicas = *gb.Spec.Replicas
	}

	image := gb.Spec.Image

	containerPorts := func() []corev1.ContainerPort {
		if len(gb.Spec.Ports) == 0 {
			return []corev1.ContainerPort{{Name: "grpc", ContainerPort: 50051}}
		}
		out := make([]corev1.ContainerPort, 0, len(gb.Spec.Ports))
		for _, p := range gb.Spec.Ports {
			out = append(out, corev1.ContainerPort{
				Name:          p.Name,
				ContainerPort: p.ContainerPort,
				Protocol:      p.Protocol,
			})
		}
		return out
	}()

	spec := appsv1.DeploymentSpec{
		Replicas: ptr.To(replicas),
		Selector: &metav1.LabelSelector{MatchLabels: lbl},
		Template: corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels: lbl,
			},
			Spec: corev1.PodSpec{
				ServiceAccountName: fmt.Sprintf("%s-sa", gb.Name),
				Containers: []corev1.Container{
					{
						Name:      "server",
						Image:     image,
						Env:       gb.Spec.Env,
						Resources: gb.Spec.Resources,
						Ports:     containerPorts,
						ReadinessProbe: &corev1.Probe{
							ProbeHandler: corev1.ProbeHandler{
								TCPSocket: &corev1.TCPSocketAction{Port: intstr.FromInt(50051)},
							},
						},
						LivenessProbe: &corev1.Probe{
							ProbeHandler: corev1.ProbeHandler{
								TCPSocket: &corev1.TCPSocketAction{Port: intstr.FromInt(50051)},
							},
						},
					},
				},
			},
		},
	}

	switch gb.Spec.UpdateStrategy {
	case apiv1alpha1.UpdateStrategyRecreate:
		spec.Strategy = appsv1.DeploymentStrategy{Type: appsv1.RecreateDeploymentStrategyType}
	default:
		spec.Strategy = appsv1.DeploymentStrategy{Type: appsv1.RollingUpdateDeploymentStrategyType}
	}

	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-deploy", gb.Name),
			Namespace: gb.Namespace,
			Labels:    lbl,
		},
		Spec: spec,
	}
}
