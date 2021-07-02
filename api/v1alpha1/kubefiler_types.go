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

// KubeFilerStatus defines the observed state of KubeFiler
type KubeFilerStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
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
