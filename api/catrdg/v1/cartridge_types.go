package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.
type Subnet struct {
	Name string `json:"name"`

	Protocol string `json:"protocol"`

	CIDRBlock string `json:"cidrBlock"`

	Gateway string `json:"gateway"`
}

type Network struct {
	// The name of the network.
	Name string `json:"name"`

	// The type of the network, e.g., "public" or "private".
	Interface string `json:"interface"`

	// The IP address specified or dhcp assigned.
	Address string `json:"address"`
}

type Host struct {
	// The name of the host.
	Name string `json:"name"`

	// The type of the host, e.g., "u1.xlarge"
	InstanceType string `json:"instanceType"`

	Disk string `json:"disk"`

	BaseImage string `json:"baseImage"`

	Networks []Network `json:"networks"`

	Labels []string `json:"labels"`

	VDIaccess string `json:"vdiAccess,omitempty"`

	Condition string `json:"condition,omitempty"`
}

// CartridgeSpec defines the desired state of Cartridge
type CartridgeSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// The name to show in the Range configuration interface
	DisplayName string `json:"display_name,omitempty"`

	// To describe the nature of the Simulation, showing in UI.
	Description string `json:"description,omitempty"`

	// Values by default and available in SimRequest to substitute variables defined in the containers manifest of a Cartridge.
	Values string `json:"values"`

	// The networking configuration for the simulation environment.
	Networks []Subnet `json:"networks"`

	// The virtual machines to be created.
	Hosts []Host `json:"hosts"`

	// The Ansible playbooks for the virtual machines configuration
	Playbooks string `json:"playbooks,omitempty"`

	// The base-64 image of a Simulation being easily identifiable.
	Logo string `json:"logo,omitempty"`
}

// CartridgeStatus defines the observed state of Cartridge
type CartridgeStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster

// Cartridge is the Schema for the cartridges API
type Cartridge struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CartridgeSpec   `json:"spec,omitempty"`
	Status CartridgeStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// CartridgeList contains a list of Cartridge
type CartridgeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Cartridge `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Cartridge{}, &CartridgeList{})
}
