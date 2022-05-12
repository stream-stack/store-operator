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
	cfg "sigs.k8s.io/controller-runtime/pkg/config/v1alpha1"
)

type SubscriptionDefault struct {
	MaxRetries         uint            `json:"maxRetries"`
	MaxRequestDuration metav1.Duration `json:"maxRequestDuration"`
	AckDuration        metav1.Duration `json:"ackDuration"`
}

type StoreDefault struct {
	Image       string          `json:"image"`
	Replicas    int32           `json:"replicas"`
	Port        string          `json:"port"`
	VolumePath  string          `json:"volumePath"`
	TopologyKey string          `json:"topologyKey"`
	GrpcTimeOut metav1.Duration `json:"grpcTimeOut"`
}
type DispatcherDefault struct {
	Image                     string          `json:"image"`
	Replicas                  int32           `json:"replicas"`
	MetricsUriFormat          string          `json:"metricsUriFormat"`
	MetricsPort               string          `json:"metricsPort"`
	Timeout                   metav1.Duration `json:"timeout"`
	BrokerSystemPartitionName string          `json:"brokerSystemPartitionName"`
}

type PublisherDefault struct {
	Image    string `json:"image"`
	Replicas int32  `json:"replicas"`
	Port     string `json:"port"`
}

type BrokerDefault struct {
	Dispatcher DispatcherDefault `json:"dispatcher"`
	Publisher  PublisherDefault  `json:"publisher"`
}
type WorkflowDefault struct {
	FinalizerName string          `json:"finalizerName"`
	RetryDuration metav1.Duration `json:"retryDuration"`
}

//+kubebuilder:object:root=true

// StreamControllerConfig is the Schema for the streamcontrollerconfigs API
type StreamControllerConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// ControllerManagerConfigurationSpec returns the contfigurations for controllers
	cfg.ControllerManagerConfigurationSpec `json:",inline"`

	Workflow WorkflowDefault `json:"workflow"`

	Subscription SubscriptionDefault `json:"subscription"`
	Store        StoreDefault        `json:"store"`
	Broker       BrokerDefault       `json:"broker"`
}

func init() {
	SchemeBuilder.Register(&StreamControllerConfig{})
}
