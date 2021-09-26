/*
Copyright 2021.

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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// KubeFilerSpec defines the desired state of KubeFiler
type KubeFilerSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Portal configuration
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength:=1
	Portal string `json:"portal,omitempty"`

	// Storage defines the type and location of the storage that backs this
	// share.
	Storage KubeFilerStorageSpec `json:"storage"`
}

// KubeFilerStorageSpec defines how storage is associated with the KubeFiler.
type KubeFilerStorageSpec struct {
	// Pvc defines PVC backed storage for this share.
	// +optional
	Pvc *KubeFilerPvcSpec `json:"pvc,omitempty"`
}

// KubeFilerPvcSpec defines how a PVC may be associated with the KubeFiler.
type KubeFilerPvcSpec struct {
	// Name of the PVC to use for the share.
	// +optional
	Name string `json:"name,omitempty"`

	// Spec defines a new, temporary, PVC to use for the share.
	// Behaves similar to the embedded PVC spec for pods.
	// +optional
	Spec *corev1.PersistentVolumeClaimSpec `json:"spec,omitempty"`
}

// KubeFilerPhase is a label for the condition of a kubefiler at the current time.
type KubeFilerPhase string

// These are the valid statuses of pods.
const (
	// KubeFilerPending means the kubefiler has been accepted by the system, but the deployment of its child
	// resources has not yet started
	KubeFilerPending KubeFilerPhase = "Pending"
	// KubeFilerDeploying means that the system is currently deploying the kubefiler's child resources
	KubeFilerDeploying KubeFilerPhase = "Deploying"
	// KubeFilerDeployed means that the system has completed deploying the kubefiler's child resources and is now waiting
	// for them to start
	KubeFilerDeployed KubeFilerPhase = "Deployed"
	// KubeFilerConfiguring means that the kubefiler's child resources are running and are being configured
	KubeFilerConfiguring KubeFilerPhase = "Configuring"
	// KubeFilerRunning means the kubefiler's child resources are configured and running as expected
	KubeFilerRunning KubeFilerPhase = "Running"
	// KubeFilerError means that one of the kubefiler's child resources is in error state
	KubeFilerError KubeFilerPhase = "Error"
	// KubeFilerUnknown means that for some reason the state of the kubefiler could not be obtained, typically due
	// to an error in communicating with the host of the filer.
	KubeFilerUnknown KubeFilerPhase = "Unknown"
)

// KubeFilerStatus defines the observed state of KubeFiler
type KubeFilerStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// The phase of a KubeFiler is a simple, high-level summary of where the KubeFiler is in its lifecycle.
	// The status of the kubefiler's child resources contain more detail about the kubefiler's status.
	// There are seven possible phase values:
	// Pending: The kubefiler has been accepted by the system, but the deployment of its child
	// resources has not yet started
	// Deploying: The system is currently deploying the kubefiler's child resources
	// Deployed: The system has completed deploying the kubefiler's child resources and is now waiting
	// for them to start
	// Configuring: The kubefiler's child resources are running and are being configured
	// Running: The kubefiler's child resources are configured and running as expected
	// Error: One of the kubefiler's child resources is in error state
	// Unknown: For some reason the state of the kubefiler could not be obtained, typically due
	// to an error in communicating with the host of the filer.
	//
	// +optional
	Phase KubeFilerPhase `json:"phase,omitempty" protobuf:"bytes,1,opt,name=phase,casttype=KubeFilerPhase"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// KubeFiler is the Schema for the kubefilers API
type KubeFiler struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KubeFilerSpec   `json:"spec,omitempty"`
	Status KubeFilerStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// KubeFilerList contains a list of KubeFiler
type KubeFilerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []KubeFiler `json:"items"`
}

func init() {
	SchemeBuilder.Register(&KubeFiler{}, &KubeFilerList{})
}
