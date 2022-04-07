package store_set_steps

import (
	"bytes"
	"embed"
	"fmt"
	v1 "github.com/stream-stack/store-operator/apis/knative/v1"
	"github.com/stream-stack/store-operator/pkg/discovery"
	v12 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/yaml"
	"testing"
	"text/template"
)

//go:embed dispatcher_sts_template.yaml dispatcher_svc_template.yaml
var a embed.FS

func TestTemplate(t *testing.T) {
	tmp, err := template.New("test").Funcs(map[string]interface{}{
		"GetStreamName":              discovery.GetStreamName,
		"GetDispatcherStsName":       discovery.GetDispatcherStsName,
		"GetDispatcherContainerPort": discovery.GetDispatcherContainerPort,
	}).ParseFS(a, "*")
	if err != nil {
		panic(err)
	}
	fmt.Println(tmp)
	broker := &v1.Broker{}
	broker.ObjectMeta.Name = "name1"
	broker.ObjectMeta.Namespace = "ns1"
	broker.Labels = map[string]string{
		"l1": "val1",
		"l2": "val2",
	}
	broker.Spec.Dispatcher.Replicas = 10
	broker.Spec.Dispatcher.Image = "img1"

	//i := make([]byte, 1024,0)
	//buffer := bytes.NewBuffer(i)
	buffer := &bytes.Buffer{}
	_ = tmp.ExecuteTemplate(buffer, "dispatcher_sts_template.yaml", broker)
	fmt.Println("渲染后结果：")
	fmt.Println(string(buffer.Bytes()))
	d := &v12.Service{}
	err = yaml.Unmarshal(buffer.Bytes(), d)
	fmt.Println(err, d)
}
