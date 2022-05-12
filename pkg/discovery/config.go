package discovery

import (
	configv1 "github.com/stream-stack/store-operator/apis/config/v1"
	"time"
)

var DispatcherManagerContainerPort = `8080`
var statisticsUriFormat = `http://%s:%s/statistics`
var defaultTimeOut = time.Second * 5
var systemBrokerPartition = "_system_broker_partition"

var PublisherContainerPort = `8080`

func DefaultConfigInit(config configv1.StreamControllerConfig) {
	DispatcherManagerContainerPort = config.Broker.Dispatcher.MetricsPort
	statisticsUriFormat = config.Broker.Dispatcher.MetricsUriFormat
	defaultTimeOut = config.Broker.Dispatcher.Timeout.Duration
	systemBrokerPartition = config.Broker.Dispatcher.BrokerSystemPartitionName
	PublisherContainerPort = config.Broker.Publisher.Port
}
