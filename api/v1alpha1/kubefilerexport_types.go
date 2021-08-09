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

// KubeFilerExportSpec defines the desired state of KubeFilerExport
type KubeFilerExportSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// KubeFiler is the name of the KubeFiler object to which this KubeFilerExport is linked
	KubeFiler string `json:"kubefiler,omitempty"`

	// Path is the path on the KubeFiler to export
	Path string `json:"path,omitempty"`

	// ReadOnly controls if this export is to be read-only or not.
	// +kubebuilder:default:=false
	// +optional
	ReadOnly bool `json:"readOnly,omitempty"`
}

// KubeFilerExportStatus defines the observed state of KubeFilerExport
type KubeFilerExportStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// KubeFilerExport is the Schema for the kubefilerexports API
type KubeFilerExport struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KubeFilerExportSpec   `json:"spec,omitempty"`
	Status KubeFilerExportStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// KubeFilerExportList contains a list of KubeFilerExport
type KubeFilerExportList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []KubeFilerExport `json:"items"`
}

func init() {
	SchemeBuilder.Register(&KubeFilerExport{}, &KubeFilerExportList{})
}
