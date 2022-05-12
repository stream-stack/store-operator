package broker_steps

import (
	configv1 "github.com/stream-stack/store-operator/apis/config/v1"
	v1 "github.com/stream-stack/store-operator/apis/knative/v1"
	"github.com/stream-stack/store-operator/controllers/knative"
)

func Register(config configv1.StreamControllerConfig) {
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
