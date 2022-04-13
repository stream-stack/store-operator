package store_set_steps

import (
	v1 "github.com/stream-stack/store-operator/apis/knative/v1"
	"github.com/stream-stack/store-operator/controllers/knative"
)

func init() {
	config := &InitConfig{
		DispatcherImage:    "ccr.ccs.tencentyun.com/stream/dispatcher:latest",
		DispatcherReplicas: 2,
		PublisherImage:     "ccr.ccs.tencentyun.com/stream/publisher:latest",
		PublisherReplicas:  2,
	}
	binding := NewRoleBinding(config)
	knative.Steps = append(knative.Steps, binding)
	v1.BrokerValidators = append(v1.BrokerValidators, binding)
	v1.BrokerDefaulters = append(v1.BrokerDefaulters, binding)

	dept := NewDispatcher(config)
	knative.Steps = append(knative.Steps, dept)
	v1.BrokerValidators = append(v1.BrokerValidators, dept)
	v1.BrokerDefaulters = append(v1.BrokerDefaulters, dept)

	publisher := NewPublisher(config)
	knative.Steps = append(knative.Steps, publisher)
	v1.BrokerValidators = append(v1.BrokerValidators, publisher)
	v1.BrokerDefaulters = append(v1.BrokerDefaulters, publisher)
}

type InitConfig struct {
	DispatcherImage    string
	DispatcherReplicas int32
	PublisherImage     string
	PublisherReplicas  int32
}
