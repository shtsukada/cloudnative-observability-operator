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

// ObservabilityConfigSpec defines the desired state of ObservabilityConfig
type ObservabilityConfigSpec struct {
	// Example field: endpoint URL of OTLP exporter
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=2048
	// +kubebuilder:validation:XValidation:rule="(self.matches('^https?://.+')) || (self.matches('^[A-Za-z0-9_.-]+:[0-9]{1,5}$'))",message="endpoint must be http(s) URL or host:port"
	Endpoint string `json:"endpoint"`

	// Example: sampling ratio percent (0-100)
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=100
	// +kubebuilder:default:=10
	SamplingPercent *int32 `json:"samplingPercent,omitempty"`

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

	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:path=observabilityconfigs,scope=Namespaced,shortName=obscfg,categories=all
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Endpoint",type=string,JSONPath=`.spec.endpoint`
// +kubebuilder:printcolumn:name="Sampling",type=integer,JSONPath=`.spec.samplingPercent`
// +kubebuilder:printcolumn:name="Metrics",type=boolean,JSONPath=`.spec.metricsEnabled`
// +kubebuilder:printcolumn:name="Phase",type=string,JSONPath=`.status.phase`
// +kubebuilder:printcolumn:name="Ready",type=string,JSONPath=`.status.conditions[?(@.type=="Ready")].status`,priority=1
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// ObservabilityConfig is the Schema for the observabilityconfigs API
type ObservabilityConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// +required
	Spec ObservabilityConfigSpec `json:"spec,omitempty"`

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
