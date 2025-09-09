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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ObservabilityConfigSpec defines the desired state of ObservabilityConfig
type ObservabilityConfigSpec struct {
	// Example field: endpoint URL of OTLP exporter
	// +kubebuilder:validation:Pattern=`^https?://`
	// +kubebuilder:validation:MaxLength=2048
	Endpoint string `json:"endpoint"`

	// Example: sampling ratio (0.0 - 1.0)
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=1
	// +kubebuilder:default:=0.1
	Sampling float64 `json:"sampling,omitempty"`

	// Enable metrics pipeline
	// +kubebuilder:default:=true
	MetricsEnabled bool `json:"metricsEnabled,omitempty"`
}

// ObservabilityConfigStatus defines the observed state of ObservabilityConfig.
type ObservabilityConfigStatus struct {
	// +kubebuilder:validation:Enum=Ready;Error;Reconciling
	// +kubebuilder:default:=Reconciling
	Phase string `json:"phase,omitempty"`

	// Reason for last transition
	// +kubebuilder:validation:MaxLength=1024
	Reason string `json:"reason,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=observabilityconfigs,scope=Namespaced,shortName=obscfg,categories=all

/*
+kubebuilder:printcolumn:name="Endpoint",type=string,JSONPath=`.spec.endpoint`
+kubebuilder:printcolumn:name="Sampling",type=string,JSONPath=`.spec.sampling`
+kubebuilder:printcolumn:name="Phase",type=string,JSONPath=`.status.phase`
+kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`
*/

// ObservabilityConfig is the Schema for the observabilityconfigs API

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
type ObservabilityConfig struct {
	metav1.TypeMeta `json:",inline"`

	// metadata is a standard object metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// spec defines the desired state of ObservabilityConfig
	// +required
	Spec ObservabilityConfigSpec `json:"spec,omitempty"`

	// status defines the observed state of ObservabilityConfig
	// +optional
	Status ObservabilityConfigStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ObservabilityConfigList contains a list of ObservabilityConfig
type ObservabilityConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ObservabilityConfig `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ObservabilityConfig{}, &ObservabilityConfigList{})
}
