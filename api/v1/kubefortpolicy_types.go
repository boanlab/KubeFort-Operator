// SPDX-License-Identifier: Apache-2.0
// Copyright 2025 BoanLab @ DKU

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

type ProcessRule struct {
	// +kubebuilder:validation:Pattern=^[^/]+$
	Name string `json:"name,omitempty"`
	// +kubebuilder:validation:Pattern=^/
	Path string `json:"path,omitempty"`
	// +kubebuilder:validation:Pattern=^/
	Dir        string        `json:"dir,omitempty"`
	Recursive  bool          `json:"recursive,omitempty"`
	FromSource []SourceMatch `json:"fromSource,omitempty"`
}

type FileRule struct {
	// +kubebuilder:validation:Pattern=^[^/]+$
	Name string `json:"name,omitempty"`
	// +kubebuilder:validation:Pattern=^/
	Path string `json:"path,omitempty"`
	// +kubebuilder:validation:Pattern=^/
	Dir        string        `json:"dir,omitempty"`
	Recursive  bool          `json:"recursive,omitempty"`
	ReadOnly   bool          `json:"readOnly,omitempty"`
	FromSource []SourceMatch `json:"fromSource,omitempty"`
}

type NetworkRule struct {
	// +kubebuilder:validation:Enum=ingress;egress
	Direction      string            `json:"direction"`
	TargetSelector map[string]string `json:"targetSelector,omitempty"`
	IPBlock        IPBlock           `json:"ipBlock,omitempty"`
	Ports          []Port            `json:"ports,omitempty"`
	FromSource     []SourceMatch     `json:"fromSource,omitempty"`
}

type IPBlock struct {
	// +kubebuilder:validation:Pattern=^[0-9.]+/[0-9]+$
	CIDR string `json:"cidr"`
	// Except is a list of CIDR ranges that should be excluded from the CIDR range specified in CIDR.
	// Each CIDR must be a valid IPv4 CIDR in the format of x.x.x.x/y where x is 0-255 and y is 0-32.
	Except []string `json:"except,omitempty"`
}

type Port struct {
	// +kubebuilder:validation:Enum=IP;TCP;UDP;ICMP
	Protocol string `json:"protocol"`
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=65535
	Port int32 `json:"port"`
}

type SourceMatch struct {
	// +kubebuilder:validation:Pattern=^[^/]+$
	Name string `json:"name"`
	// +kubebuilder:validation:Pattern=^/
	Path string `json:"path"`
}

// KubeFortPolicySpec defines the desired state of KubeFortPolicy.
type KubeFortPolicySpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	Selector map[string]string `json:"selector,omitempty"`

	Process []ProcessRule `json:"process,omitempty"`
	File    []FileRule    `json:"file,omitempty"`
	Network []NetworkRule `json:"network,omitempty"`

	// +kubebuilder:validation:Enum=Allow;Audit;Block
	Action string `json:"action"`
}

// KubeFortPolicyStatus defines the observed state of KubeFortPolicy.
type KubeFortPolicyStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	PolicyStatus string `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=kfp
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`
// +kubebuilder:printcolumn:name="Action",type=string,JSONPath=`.spec.action`,priority=10
// +kubebuilder:printcolumn:name="Selector",type=string,JSONPath=`.spec.selector.matchLabels`,priority=10

// KubeFortPolicy is the Schema for the kubefortpolicies API.
type KubeFortPolicy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KubeFortPolicySpec   `json:"spec,omitempty"`
	Status KubeFortPolicyStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// KubeFortPolicyList contains a list of KubeFortPolicy.
type KubeFortPolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []KubeFortPolicy `json:"items"`
}

func init() {
	SchemeBuilder.Register(&KubeFortPolicy{}, &KubeFortPolicyList{})
}
