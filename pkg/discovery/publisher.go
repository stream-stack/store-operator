package discovery

import (
	"fmt"
	v1 "github.com/stream-stack/store-operator/apis/knative/v1"
)

func GetPublisherName(b *v1.Broker) string {
	return fmt.Sprintf(`%s-publisher`, b.Name)
}

func GetPublisherManagerContainerPort() string {
	return PublisherContainerPort
}
