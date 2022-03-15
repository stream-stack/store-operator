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
	v1 "k8s.io/api/apps/v1"
	v12 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

type BrokerPartition struct {
	//按照 数量，数据量，时间 三种策略进行分区
	//Counter  BrokerPartitionCounter  `json:"counter,omitempty"`
	//DataSize BrokerPartitionDataSize `json:"dataSize,omitempty"`
	//Timer    BrokerPartitionTimer    `json:"timer,omitempty"`
	Size uint64 `json:"size,omitempty"`
}

type DispatcherSpec struct {
	Replicas int32  `json:"replicas,omitempty"`
	Image    string `json:"image,omitempty"`
}

type PublisherSpec struct {
	Replicas int32  `json:"replicas,omitempty"`
	Image    string `json:"image,omitempty"`
}

// BrokerSpec defines the desired state of Broker
type BrokerSpec struct {
	Selector   *metav1.LabelSelector `json:"selector,omitempty"`
	Partition  BrokerPartition       `json:"partition,omitempty"`
	Dispatcher DispatcherSpec        `json:"dispatcher,omitempty"`
	Publisher  PublisherSpec         `json:"publisher,omitempty"`
}

type DispatcherStatus struct {
	Sts     v1.StatefulSetStatus `json:"sts,omitempty"`
	SvcName string               `json:"svcName,omitempty"`
}

type PublisherStatus struct {
	Sts     v1.StatefulSetStatus `json:"sts,omitempty"`
	SvcName string               `json:"svcName,omitempty"`
}

type ServiceStatus struct {
	v12.ServiceStatus `json:",inline"`
	Name              string `json:"name,omitempty"`
}

// BrokerStatus defines the observed state of Broker
type BrokerStatus struct {
	Dispatcher DispatcherStatus `json:"dispatcher,omitempty"`
	Publisher  PublisherStatus  `json:"publisher,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Broker is the Schema for the brokers API
type Broker struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   BrokerSpec   `json:"spec,omitempty"`
	Status BrokerStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// BrokerList contains a list of Broker
type BrokerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Broker `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Broker{}, &BrokerList{})
}
