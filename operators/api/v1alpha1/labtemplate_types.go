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
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type LabType string
type EnvironmentType string

const (
	TypeGUI LabType = "GUI"
	TypeCLI LabType = "CLI"
	ClassContainer EnvironmentType = "Container"
	ClassVM EnvironmentType = "VirtualMachine"
)



// LabTemplateSpec defines the desired state of LabTemplate
type LabTemplateSpec struct {
	CourseName  string                        `json:"courseName,omitempty"`
	LabName     string                        `json:"labName,omitempty"`
	LabNum      resource.Quantity             `json:"labNum,omitempty"`
	Description string                        `json:"description,omitempty"`
	LabEnvironmentList []LabEnvironment
	// +kubebuilder:validation:Enum="GUI";"CLI"

}

// LabTemplateStatus defines the observed state of LabTemplate
type LabTemplateStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make lab-template" to regenerate code after modifying this file
}

type LabEnvironment struct {
	Name string  `json:"Name,omitempty"`
	LabType `json:"LabType,omitempty"`
    Resources resource.Quantity `json:"resources,omitempty"`
	Type  EnvironmentType `json:"type"`
	Persistent bool `json:"persistent"`
	Image string `json:"image"`
}


// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName="labt"

// LabTemplate is the Schema for the labtemplates API
type LabTemplate struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   LabTemplateSpec   `json:"spec,omitempty"`
	Status LabTemplateStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// LabTemplateList contains a list of LabTemplate
type LabTemplateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []LabTemplate `json:"items"`
}

func init() {
	SchemeBuilder.Register(&LabTemplate{}, &LabTemplateList{})
}
