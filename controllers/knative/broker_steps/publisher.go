package store_set_steps

import (
	"bytes"
	"embed"
	"fmt"
	v14 "github.com/stream-stack/store-operator/apis/knative/v1"
	"github.com/stream-stack/store-operator/pkg/base"
	v1 "k8s.io/api/apps/v1"
	v12 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/yaml"
	"reflect"
	"text/template"
)

//go:embed publisher_sts_template.yaml publisher_svc_template.yaml
var publisherTemplateFs embed.FS

var publisherYamlTemplate *template.Template

func init() {
	var err error
	publisherYamlTemplate, err = template.ParseFS(publisherTemplateFs, "*")
	if err != nil {
		panic(err)
	}
}

func NewPublisher(config *InitConfig) *base.Step {
	sts := &base.Step{
		Name: fmt.Sprintf(`publisher-sts`),
		GetObj: func() base.StepObject {
			return &v1.StatefulSet{}
		},
		Render: func(t base.StepObject) (base.StepObject, error) {
			c := t.(*v14.Broker)
			buffer := &bytes.Buffer{}
			err := publisherYamlTemplate.ExecuteTemplate(buffer, "publisher_sts_template.yaml", c)
			if err != nil {
				return nil, err
			}
			fmt.Println("渲染后结果：")
			fmt.Println(string(buffer.Bytes()))

			d := &v1.StatefulSet{}
			err = yaml.Unmarshal(buffer.Bytes(), d)
			if err != nil {
				return nil, err
			}

			return d, nil
		},
		SetStatus: func(owner base.StepObject, target, now base.StepObject) (needUpdate bool, updateObject base.StepObject, err error) {
			c := owner.(*v14.Broker)
			o := now.(*v1.StatefulSet)
			c.Status.Publisher.Sts = o.Status

			t := target.(*v1.StatefulSet)
			if !reflect.DeepEqual(t.Spec, o.Spec) {
				o.Spec = t.Spec
				return true, o, nil
			}

			return false, now, nil
		},
		Next: func(ctx *base.StepContext) (bool, error) {
			broker := ctx.StepObject.(*v14.Broker)
			//todo:将符合selector的subscription地址推送到publisher
			if broker.Status.Publisher.Sts.ReadyReplicas == broker.Spec.Publisher.Replicas {
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
	}
	svc := &base.Step{
		Name: fmt.Sprintf(`publisher-svc`),
		GetObj: func() base.StepObject {
			return &v12.Service{}
		},
		Render: func(t base.StepObject) (base.StepObject, error) {
			c := t.(*v14.Broker)
			buffer := &bytes.Buffer{}
			err := publisherYamlTemplate.ExecuteTemplate(buffer, "publisher_svc_template.yaml", c)
			if err != nil {
				return nil, err
			}
			fmt.Println("渲染后结果：")
			fmt.Println(string(buffer.Bytes()))

			d := &v12.Service{}
			err = yaml.Unmarshal(buffer.Bytes(), d)
			if err != nil {
				return nil, err
			}

			return d, nil
		},
		SetStatus: func(owner base.StepObject, target, now base.StepObject) (needUpdate bool, updateObject base.StepObject, err error) {
			c := owner.(*v14.Broker)
			o := now.(*v12.Service)
			c.Status.Publisher.Svc = o.Status

			t := target.(*v12.Service)
			if !reflect.DeepEqual(t.Spec.Selector, o.Spec.Selector) {
				o.Spec.Selector = t.Spec.Selector
				return true, o, nil
			}
			if !reflect.DeepEqual(t.Spec.Ports, o.Spec.Ports) {
				o.Spec.Ports = t.Spec.Ports
				return true, o, nil
			}
			if !reflect.DeepEqual(t.Spec.Type, o.Spec.Type) {
				o.Spec.Type = t.Spec.Type
				return true, o, nil
			}
			return false, now, nil
		},
		Next: func(ctx *base.StepContext) (bool, error) {
			return true, nil
		},
	}
	return &base.Step{
		Name: "publisher",
		Sub:  []*base.Step{svc, sts},
	}
}
