/*
Copyright 2020 pc.

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

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	rbacv1 "k8s.io/api/rbac/v1"

)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// UserIdentitySpec defines the desired state of UserIdentity
type UserIdentitySpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// RoleRef is the target ClusterRole reference
	RoleRef rbacv1.RoleRef `json:"roleRef,omitempty"`
}

// UserIdentityStatus defines the observed state of UserIdentity
type UserIdentityStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +kubebuilder:object:root=true

// UserIdentity is the Schema for the useridentities API
type UserIdentity struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   UserIdentitySpec   `json:"spec,omitempty"`
	Status UserIdentityStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// UserIdentityList contains a list of UserIdentity
type UserIdentityList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []UserIdentity `json:"items"`
}

func init() {
	SchemeBuilder.Register(&UserIdentity{}, &UserIdentityList{})
}
