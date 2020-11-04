/*


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

// WorkspaceUserRole is an enum for the role of a user in a workspace
// +kubebuilder:validation:Enum=Admin;Basic
type WorkspaceUserRole string

const (
	// Admin allows to interact with all VMs of a workspace
	Admin WorkspaceUserRole = "Admin"
	// Basic allows to interact with owned vms
	Basic WorkspaceUserRole = "Basic"
)

// UserWorkspaceData contains the info of the workspaces related to a user
type UserWorkspaceData struct {
	WorkspaceURL string            `json:"workspaceURL"`
	GroupNumber  int               `json:"groupNumber"`
	Role         WorkspaceUserRole `json:"role"`
}

// TenantSpec defines the desired state of Tenant
type TenantSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	Name string `json:"name"`

	SurName string `json:"surname"`

	// +kubebuilder:validation:MinLength=8
	ID string `json:"ID"`

	// +kubebuilder:validation:Pattern="^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$"
	Email string `json:"email"`

	// list of workspaces the user is in
	Workspaces []UserWorkspaceData `json:"workspaces,omitempty"`

	// public keys of user
	PublicKeys []string `json:"publicKeys,omitempty"`
}

// NameCreated contains info about the status of a resource
type NameCreated struct {
	Name    string `json:"name"`
	Created bool   `json:"created"`
}

// SubscriptionStatus is an enum for the status of a subscription to a service
// +kubebuilder:validation:Enum=Ok;Pending;Failed
type SubscriptionStatus string

const (
	// Ok -> the subscription was successful
	Ok SubscriptionStatus = "Ok"
	// Pending -> the subscription is in process
	Pending SubscriptionStatus = "Pending"
	// Failed -> the subscription has failed
	Failed SubscriptionStatus = "Failed"
)

// TenantStatus defines the observed state of Tenant
type TenantStatus struct {
	PersonalNamespace NameCreated                   `json:"personalNamespace"`
	NamespaceSandbox  NameCreated                   `json:"namespaceSandbox"`
	Subscriptions     map[string]SubscriptionStatus `json:"subscriptionStatus"`
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +kubebuilder:object:root=true
// +kubebuilder:printcolumn:name="Name",type=string,JSONPath=`.spec.name`
// +kubebuilder:printcolumn:name="Surname",type=string,JSONPath=`.spec.surname`
// +kubebuilder:printcolumn:name="Email",type=string,JSONPath=`.spec.email`
// +kubebuilder:printcolumn:name="ID",type=string,JSONPath=`.spec.ID`

// Tenant is the Schema for the tenants API
type Tenant struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TenantSpec   `json:"spec,omitempty"`
	Status TenantStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// TenantList contains a list of Tenant
type TenantList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Tenant `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Tenant{}, &TenantList{})
}
