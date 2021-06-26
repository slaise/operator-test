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

package v3

import (
	"github.com/operator-framework/operator-lib/status"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.
type TemplateObject struct {
	// +kubebuilder:pruning:PreserveUnknownFields
	unstructured.Unstructured `json:",inline"`
}

// UserIdentityV3Spec defines the desired state of UserIdentityV3
type UserIdentityV3Spec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Template is a list of resources to instantiate per repository in Governator
	Template []TemplateObject `json:"template,omitempty"`

	// Name is a test field for webhook validation
	// +kubebuilder:validation:Required
	name string `json:"name"`
}

// UserIdentityV3Status defines the observed state of UserIdentityV3
type UserIdentityV3Status struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Conditions is the list of error conditions for this resource
	Conditions status.Conditions `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true

// UserIdentityV3 is the Schema for the useridentityv3s API
type UserIdentityV3 struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   UserIdentityV3Spec   `json:"spec,omitempty"`
	Status UserIdentityV3Status `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// UserIdentityV3List contains a list of UserIdentityV3
type UserIdentityV3List struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []UserIdentityV3 `json:"items"`
}

func init() {
	SchemeBuilder.Register(&UserIdentityV3{}, &UserIdentityV3List{})
}

func (o *UserIdentityV3) GetConditions() *status.Conditions {
	return &o.Status.Conditions
}
