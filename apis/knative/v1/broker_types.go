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
	"fmt"
	"github.com/stream-stack/common/protocol/dispatcher"
	"github.com/stream-stack/common/protocol/operator"
	v1 "k8s.io/api/apps/v1"
	v12 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"math/rand"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

type BrokerPartitionCounter struct {
	Count uint64 `json:"count"`
}

func (c *BrokerPartitionCounter) AllocatePartition(statistics *dispatcher.Statistics, sets []*operator.StoreSet) (*operator.StoreSet, uint64, error) {
	if statistics.PartitionCount == 0 {
		intn := rand.Intn(len(sets))
		set := sets[intn]

		return set, 1, nil
	}
	intn := rand.Intn(len(sets))
	set := sets[intn]

	//如果当前eventId为10,count为2,PartitionCount为3
	//则下一个分片begin 10+2,PartitionCount为4
	//
	if statistics.MaxEvent > statistics.PartitionCount*c.Count {
		return set, statistics.MaxEvent + c.Count, nil
	}

	return nil, 1, nil
}

type BrokerPartition struct {
	//按照 数量，数据量，时间 三种策略进行分区
	Counter *BrokerPartitionCounter `json:"counter,omitempty"`
	//TODO:实现其他策略
	//DataSize BrokerPartitionDataSize `json:"dataSize,omitempty"`
	//Timer    BrokerPartitionTimer    `json:"timer,omitempty"`
}

func (in *BrokerPartition) AllocatePartition(statistics *dispatcher.Statistics, sets []*operator.StoreSet) (*operator.StoreSet, uint64, error) {
	if in.Counter != nil {
		return in.Counter.AllocatePartition(statistics, sets)
	}
	return nil, 0, fmt.Errorf("no partition strategy")
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
	Uuid       string                `json:"uuid,omitempty"`
	Selector   *metav1.LabelSelector `json:"selector,omitempty"`
	Partition  BrokerPartition       `json:"partition,omitempty"`
	Dispatcher DispatcherSpec        `json:"dispatcher,omitempty"`
	Publisher  PublisherSpec         `json:"publisher,omitempty"`
}

type DispatcherStatus struct {
	WorkloadStatus v1.DeploymentStatus `json:"workloadStatus,omitempty"`
	SvcName        string              `json:"svcName,omitempty"`
}

type PublisherStatus struct {
	WorkloadStatus v1.DeploymentStatus `json:"workloadStatus,omitempty"`
}

type ServiceStatus struct {
	v12.ServiceStatus `json:",inline"`
	Name              string `json:"name,omitempty"`
}

// BrokerStatus defines the observed state of Broker
type BrokerStatus struct {
	Uuid       string           `json:"uuid,omitempty"`
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
