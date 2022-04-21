package discovery

import (
	"fmt"
	_ "github.com/Jille/grpc-multi-resolver"
	"github.com/sirupsen/logrus"
	v1 "github.com/stream-stack/store-operator/apis/knative/v1"
	_ "google.golang.org/grpc/health"
	"k8s.io/apimachinery/pkg/util/json"
)

func GetStreamName(b *v1.Broker) string {
	return fmt.Sprintf("%s-%s-%s", b.Namespace, b.Name, b.Status.Uuid)
}
func GetSelector(b *v1.Broker) string {
	marshal, err := json.Marshal(b.Spec.Selector)
	if err != nil {
		logrus.Warnf("marshal selector error:%v", err)
	}
	return string(marshal)
}

func GetDispatcherDeptName(b *v1.Broker) string {
	return fmt.Sprintf(`%s-dispatcher`, b.Name)
}

const DispatcherManagerContainerPort = `8080`

func GetDispatcherManagerContainerPort() string {
	return DispatcherManagerContainerPort
}
