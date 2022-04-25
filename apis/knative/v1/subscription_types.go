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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

type SubscriberDelivery struct {
	//TODO:定义发送消息的方式
	// +optional
	//DeadLetterPolicy *DeadLetterPolicy `json:"deadLetterPolicy,omitempty"`
	// +optional
	//RetryPolicy *RetryPolicy `json:"retryPolicy,omitempty"`
	// +optional
	//BackoffPolicy *BackoffPolicy `json:"backoffPolicy,omitempty"`
}

type SubscriberUri struct {
	Uri      string `json:"uri,omitempty"`
	Protocol string `json:"protocol,omitempty"`
}

type SubscriberService struct {
	Name      string `json:"name,omitempty"`
	Namespace string `json:"namespace,omitempty"`
	Port      int32  `json:"port,omitempty"`
	Path      string `json:"path,omitempty"`
}

type SubscriberPod struct {
	Namespace string                `json:"namespace,omitempty"`
	Selector  *metav1.LabelSelector `json:"selector,omitempty"`
	Protocol  string                `json:"protocol,omitempty"`
	Port      int32                 `json:"port,omitempty"`
	Path      string                `json:"path,omitempty"`
}

type SubscriberFilter struct {
	//TODO:定义过滤字段
}

type Subscriber struct {
	Uri     *SubscriberUri     `json:"uri,omitempty"`
	Service *SubscriberService `json:"service,omitempty"`
	Pod     *SubscriberPod     `json:"pod,omitempty"`
}

// SubscriptionSpec defines the desired state of Subscription
type SubscriptionSpec struct {
	Broker     string              `json:"broker"`
	Delivery   *SubscriberDelivery `json:"delivery,omitempty"`
	Filter     *SubscriberFilter   `json:"filter,omitempty"`
	Subscriber Subscriber          `json:"subscriber,omitempty"`
}

// SubscriptionStatus defines the observed state of Subscription
type SubscriptionStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Subscription is the Schema for the subscriptions API
type Subscription struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SubscriptionSpec   `json:"spec,omitempty"`
	Status SubscriptionStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// SubscriptionList contains a list of Subscription
type SubscriptionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Subscription `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Subscription{}, &SubscriptionList{})
}
