package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.
type Network struct {
	Name string `json:"name"`
	Interface string `json:"interface,omitempty"`
	Namespace string `json:"namespace,omitempty"`
}

// CsrxSpec defines the desired state of Csrx
// +k8s:openapi-gen=true
type CsrxSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book.kubebuilder.io/beyond_basics/generating_crd.html
	InitImage string `json:"initImage"`
	InitImagePullPolicy string `json:"initImagePullPolicy,omitempty"`
	CsrxImage string `json:"csrxImage"`
	CsrxImagePullPolicy string `json:"csrxImagePullPolicy,omitempty"`
	ImagePullSecrets []string `json:"imagePullSecrets,omitempty"`
	Networks []Network `json:"networks"`
}

// CsrxStatus defines the observed state of Csrx
// +k8s:openapi-gen=true
type CsrxStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book.kubebuilder.io/beyond_basics/generating_crd.html
	Nodes []string `json:"nodes"`
	Prefix string `json:"prefix"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Csrx is the Schema for the csrxes API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
type Csrx struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CsrxSpec   `json:"spec,omitempty"`
	Status CsrxStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// CsrxList contains a list of Csrx
type CsrxList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Csrx `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Csrx{}, &CsrxList{})
}
