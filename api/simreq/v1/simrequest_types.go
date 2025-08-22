package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// SimRequestSpec defines the desired state of SimRequest
type SimRequestSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// The namespace where the cartridge resources will be deployed and the simulation will run.
	Grid string `json:"grid"`

	// The idetifier or description of a Simulation to be launched.
	Simulation string `json:"simulation,omitempty"`

	// The name of the requested cartridge to activate
	CartridgeName string `json:"cartridge_name"`

	// Values in yaml are used to substitute variables defined in the containers manifest of a Cartridge.
	Values string `json:"values,omitempty"`
}

// SimRequestStatus defines the observed state of SimRequest
type SimRequestStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	StartTime     string `json:"start_time,omitempty"`
	AppliedValues string `json:"applied_values,omitempty"`
	Result        string `json:"result"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// SimRequest is the Schema for the simrequests API
type SimRequest struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SimRequestSpec   `json:"spec,omitempty"`
	Status SimRequestStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// SimRequestList contains a list of SimRequest
type SimRequestList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SimRequest `json:"items"`
}

func init() {
	SchemeBuilder.Register(&SimRequest{}, &SimRequestList{})
}
