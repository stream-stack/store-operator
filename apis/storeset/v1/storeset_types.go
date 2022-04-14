/*
Copyright 2022.

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
	v12 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

type LocalPvSpec struct {
	NodeAffinity      *v1.VolumeNodeAffinity `json:"nodeAffinity,omitempty"`
	Capacity          resource.Quantity      `json:"capacity"`
	LocalVolumeSource *v1.LocalVolumeSource  `json:"source,omitempty"`
}

type StoreStatefulSetSpec struct {
	Replicas *int32 `json:"replicas,omitempty"`
	Image    string `json:"image,omitempty"`
	Port     string `json:"port,omitempty"`
}

// StoreSetSpec defines the desired state of StoreSet
type StoreSetSpec struct {
	Volume LocalPvSpec          `json:"volume,omitempty"`
	Store  StoreStatefulSetSpec `json:"store,omitempty"`
}

type LocalPvStatus struct {
	Name   string                    `json:"name"`
	Status v1.PersistentVolumeStatus `json:",inline"`
}

type StoreStatefulSetStatus struct {
	WorkloadName string                `json:"workloadName"`
	ServiceName  string                `json:"serviceName"`
	Workload     v12.StatefulSetStatus `json:"workload"`
	Service      v1.ServiceStatus      `json:"service"`
}

type PublisherDeptStatus struct {
	Name   string               `json:"name"`
	Status v12.DeploymentStatus `json:",inline"`
}

// StoreSetStatus defines the observed state of StoreSet
type StoreSetStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	VolumeStatus LocalPvStatus          `json:"volume"`
	StoreStatus  StoreStatefulSetStatus `json:"store"`
	Status       string                 `json:"status"`
}

const StoreSetStatusReady = `ready`
const StoreSetStatusPVCreating = `PVCreating`
const StoreSetStatusStoreSetSvcCreating = `StoreSetSvcCreating`
const StoreSetStatusStoreSetStsCreating = `StoreSetStsCreating`

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// StoreSet is the Schema for the storesets API
type StoreSet struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   StoreSetSpec   `json:"spec,omitempty"`
	Status StoreSetStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// StoreSetList contains a list of StoreSet
type StoreSetList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []StoreSet `json:"items"`
}

func init() {
	SchemeBuilder.Register(&StoreSet{}, &StoreSetList{})
}
