package store_set_steps

import (
	v1 "github.com/stream-stack/store-operator/apis/knative/v1"
	"github.com/stream-stack/store-operator/controllers/knative"
)

func init() {
	config := &InitConfig{
		BrokerImage:       "ccr.ccs.tencentyun.com/stream/broker:latest",
		BrokerReplicas:    2,
		PublisherImage:    "ccr.ccs.tencentyun.com/stream/publisher:latest",
		PublisherReplicas: 2,
	}
	dept := NewDispatcher(config)
	knative.Steps = append(knative.Steps, dept)
	v1.BrokerValidators = append(v1.BrokerValidators, dept)
	v1.BrokerDefaulters = append(v1.BrokerDefaulters, dept)

	publisher := NewPublisher(config)
	knative.Steps = append(knative.Steps, dept)
	v1.BrokerValidators = append(v1.BrokerValidators, publisher)
	v1.BrokerDefaulters = append(v1.BrokerDefaulters, publisher)
}

type InitConfig struct {
	BrokerImage       string
	BrokerReplicas    int32
	PublisherImage    string
	PublisherReplicas int32
}
