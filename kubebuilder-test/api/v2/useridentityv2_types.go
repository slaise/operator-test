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

package v2

import (
	"github.com/operator-framework/operator-lib/status"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// UserIdentityV2Spec defines the desired state of UserIdentityV2
type UserIdentityV2Spec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// RoleRef is the target ClusterRole reference
	RoleRef rbacv1.RoleRef `json:"roleRef,omitempty"`
}

// UserIdentityV2Status defines the observed state of UserIdentityV2
type UserIdentityV2Status struct {

	// Conditions is the list of error conditions for this resource
	Conditions status.Conditions `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true

// UserIdentityV2 is the Schema for the useridentityv2s API
type UserIdentityV2 struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   UserIdentityV2Spec   `json:"spec,omitempty"`
	Status UserIdentityV2Status `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// UserIdentityV2List contains a list of UserIdentityV2
type UserIdentityV2List struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []UserIdentityV2 `json:"items"`
}

func init() {
	SchemeBuilder.Register(&UserIdentityV2{}, &UserIdentityV2List{})
}

func (o *UserIdentityV2) GetConditions() *status.Conditions {
	return &o.Status.Conditions
}
