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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// KubeFilerPortalSpec defines the desired state of KubeFilerPortal
type KubeFilerPortalSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Hostname or IP address of the Portal
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength:=1
	Address string `json:"foo,omitempty"`

	// Credentials for the Portal
	// +kubebuilder:validation:Required
	Credentials *KubeFilerPortalCredentialsSpec `json:"credentials,omitempty"`

	// Always trust the credentials provided by the Portal
	// +kubebuilder:default:=false
	// +optional
	Trust bool `json:"trust,omitempty"`
}

type KubeFilerPortalCredentialsSpec struct {
	// Secret identifies the name of the secret storing username and password keys
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength:=1
	Secret string `json:"secret,omitempty"`

	// UsernameKey identifies the key within the secret that stores the Username
	// +kubebuilder:default:=username
	// +optional
	UsernameKey string `json:"username_key,omitempty"`

	// PasswordKey identifies the key within the secret that stores the Password
	// +kubebuilder:default:=password
	// +optional
	PasswordKey string `json:"password_key,omitempty"`
}

// KubeFilerPortalStatus defines the observed state of KubeFilerPortal
type KubeFilerPortalStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// KubeFilerPortal is the Schema for the kubefilerportals API
type KubeFilerPortal struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KubeFilerPortalSpec   `json:"spec,omitempty"`
	Status KubeFilerPortalStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// KubeFilerPortalList contains a list of KubeFilerPortal
type KubeFilerPortalList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []KubeFilerPortal `json:"items"`
}

func init() {
	SchemeBuilder.Register(&KubeFilerPortal{}, &KubeFilerPortalList{})
}
