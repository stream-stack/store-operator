package store_set_steps

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	v14 "github.com/stream-stack/store-operator/apis/knative/v1"
	"github.com/stream-stack/store-operator/pkg/base"
	"github.com/stream-stack/store-operator/pkg/discovery"
	v1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/util/yaml"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"text/template"
)

//go:embed publisher_dept_template.yaml
var publisherTemplateFs embed.FS

var publisherYamlTemplate *template.Template

func init() {
	var err error
	publisherYamlTemplate, err = template.New("publisher").Funcs(map[string]interface{}{
		"GetSelector":                      discovery.GetSelector,
		"GetStreamName":                    discovery.GetStreamName,
		"GetPublisherName":                 discovery.GetPublisherName,
		"GetPublisherManagerContainerPort": discovery.GetPublisherManagerContainerPort,
	}).ParseFS(publisherTemplateFs, "*")
	if err != nil {
		panic(err)
	}
}

func NewPublisher(config *InitConfig) *base.Step {
	sts := &base.Step{
		Name: fmt.Sprintf(`publisher-dept`),
		GetObj: func() base.StepObject {
			return &v1.Deployment{}
		},
		Render: func(t base.StepObject) (base.StepObject, error) {
			c := t.(*v14.Broker)
			buffer := &bytes.Buffer{}
			err := publisherYamlTemplate.ExecuteTemplate(buffer, "publisher_dept_template.yaml", c)
			if err != nil {
				return nil, err
			}

			d := &v1.Deployment{}
			err = yaml.Unmarshal(buffer.Bytes(), d)
			if err != nil {
				return nil, err
			}

			return d, nil
		},
		SetStatus: func(owner base.StepObject, target, now base.StepObject) (needUpdate bool, updateObject base.StepObject, err error) {
			c := owner.(*v14.Broker)
			o := now.(*v1.Deployment)
			c.Status.Publisher.WorkloadStatus = o.Status

			t := target.(*v1.Deployment)
			if !reflect.DeepEqual(t.Spec, o.Spec) {
				o.Spec = t.Spec
				return true, o, nil
			}

			return false, now, nil
		},
		Next: func(ctx *base.StepContext) (bool, error) {
			broker := ctx.StepObject.(*v14.Broker)
			if broker.Status.Publisher.WorkloadStatus.ReadyReplicas == broker.Spec.Publisher.Replicas {
				return true, nil
			}

			return false, nil
		},
		SetDefault: func(t base.StepObject) {
			c := t.(*v14.Broker)
			if len(c.Spec.Publisher.Image) == 0 {
				c.Spec.Publisher.Image = config.PublisherImage
			}
			if c.Spec.Publisher.Replicas <= 0 {
				c.Spec.Publisher.Replicas = config.PublisherReplicas
			}
			//TODO:partition default value set
		},
		Del: func(ctx context.Context, c base.StepObject, client client.Client) error {
			//t := c.(*v14.Broker)
			//discovery.DeletePublisherConn(t)
			return nil
		},
	}
	return &base.Step{
		Name: "publisher",
		Sub:  []*base.Step{sts},
	}
}
