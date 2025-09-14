// SPDX-License-Identifier: Apache-2.0

// +kubebuilder:object:generate=true
// +groupName=observability.shtsukada.dev

package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type UpdateStrategyType string

const (
	UpdateStrategyRollingUpdate UpdateStrategyType = "RollingUpdate"
	UpdateStrategyRecreate      UpdateStrategyType = "Recreate"

	ConditionReady       string = "Ready"
	ConditionProgressing string = "Progressing"
	ConditionDegraded    string = "Degraded"
)

type PortSpec struct {
	// +optional
	Name string `json:"name,omitempty"`

	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=65535
	ContainerPort int32 `json:"containerPort"`

	// +kubebuilder:validation:Enum=TCP;UDP;SCTP
	// +kubebuilder:default:=TCP
	Protocol corev1.Protocol `json:"protocol,omitempty"`

	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=65535
	// +optional
	ServicePort *int32 `json:"servicePort,omitempty"`
}

type OTLPEndpoint struct {
	// e.g. "otel-collector.monitoring.svc:4317"
	// +kubebuilder:validation:MinLength=1
	Endpoint string `json:"endpoint"`

	// +kubebuilder:default:=false
	// +optional
	Insecure *bool `json:"insecure,omitempty"`

	// +optional
	Headers map[string]string `json:"headers,omitempty"`

	// +optional
	Timeout *metav1.Duration `json:"timeout,omitempty"`
}

type GrpcBurnerSpec struct {
	// +kubebuilder:validation:MinLength=1
	Image string `json:"image"`

	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:default:=1
	Replicas *int32 `json:"replicas,omitempty"`

	// +optional
	Env []corev1.EnvVar `json:"env,omitempty"`

	// +optional
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`

	// +kubebuilder:validation:MinItems=1
	Ports []PortSpec `json:"ports"`

	// +optional
	OTLPEndpoint *OTLPEndpoint `json:"otlpEndpoint,omitempty"`

	// +kubebuilder:validation:Enum=RollingUpdate;Recreate
	// +kubebuilder:default:=RollingUpdate
	UpdateStrategy UpdateStrategyType `json:"updateStrategy,omitempty"`
}

type GrpcBurnerStatus struct {
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// +optional
	ReadyReplicas int32 `json:"readyReplicas,omitempty"`

	// +optional
	Endpoint string `json:"endpoint,omitempty"`

	// +optional
	Phase string `json:"phase,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=gb,singular=grpcburner,scope=Namespaced
// +kubebuilder:printcolumn:name="Phase",type=string,JSONPath=`.status.phase`,description="Summary phase"
// +kubebuilder:printcolumn:name="Ready",type=integer,JSONPath=`.status.readyReplicas`,description="Ready replicas"
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`
type GrpcBurner struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GrpcBurnerSpec   `json:"spec,omitempty"`
	Status GrpcBurnerStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
type GrpcBurnerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []GrpcBurner `json:"items"`
}

func init() {
	SchemeBuilder.Register(&GrpcBurner{}, &GrpcBurnerList{})
}
