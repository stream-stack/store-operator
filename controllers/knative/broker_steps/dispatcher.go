package store_set_steps

import (
	"bytes"
	"embed"
	_ "embed"
	"fmt"
	v14 "github.com/stream-stack/store-operator/apis/knative/v1"
	v15 "github.com/stream-stack/store-operator/apis/storeset/v1"
	"github.com/stream-stack/store-operator/pkg/base"
	v1 "k8s.io/api/apps/v1"
	v12 "k8s.io/api/core/v1"
	v13 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/yaml"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"text/template"
)

//go:embed dispatcher_sts_template.yaml dispatcher_svc_template.yaml
var templateFs embed.FS

var yamlTemplate *template.Template

func init() {
	var err error
	yamlTemplate, err = template.ParseFS(templateFs, "*")
	if err != nil {
		panic(err)
	}
}

func NewDispatcher(config *InitConfig) *base.Step {
	sts := &base.Step{
		Name: fmt.Sprintf(`dispatcher-sts`),
		GetObj: func() base.StepObject {
			return &v1.StatefulSet{}
		},
		Render: func(t base.StepObject) (base.StepObject, error) {
			c := t.(*v14.Broker)
			buffer := &bytes.Buffer{}
			err := yamlTemplate.ExecuteTemplate(buffer, "dispatcher_sts_template.yaml", c)
			if err != nil {
				return nil, err
			}

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
			c.Status.Dispatcher.Sts = o.Status

			t := target.(*v1.StatefulSet)
			if !reflect.DeepEqual(t.Spec, o.Spec) {
				o.Spec = t.Spec
				return true, o, nil
			}

			return false, now, nil
		},
		Next: func(ctx *base.StepContext) (bool, error) {
			broker := ctx.StepObject.(*v14.Broker)
			if broker.Status.Dispatcher.Sts.ReadyReplicas != broker.Spec.Dispatcher.Replicas {
				return false, nil
			}
			list := &v15.StoreSetList{}
			selectorMap, err := v13.LabelSelectorAsMap(broker.Spec.Selector)
			if err != nil {
				return false, err
			}
			err = ctx.GetClient().List(ctx.Context, list, client.MatchingLabels(selectorMap))
			if err != nil {
				return false, err
			}

			//todo:将符合selector的storeset地址推送到dispatcher

			return true, nil
		},
		SetDefault: func(t base.StepObject) {
			c := t.(*v14.Broker)
			if len(c.Spec.Dispatcher.Image) == 0 {
				c.Spec.Dispatcher.Image = config.DispatcherImage
			}
			if c.Spec.Dispatcher.Replicas <= 0 {
				c.Spec.Dispatcher.Replicas = config.DispatcherReplicas
			}
			//TODO:partition default value set
		},
	}
	svc := &base.Step{
		Name: fmt.Sprintf(`dispatcher-svc`),
		GetObj: func() base.StepObject {
			return &v12.Service{}
		},
		Render: func(t base.StepObject) (base.StepObject, error) {
			c := t.(*v14.Broker)
			buffer := &bytes.Buffer{}
			if err := yamlTemplate.ExecuteTemplate(buffer, "dispatcher_svc_template.yaml", c); err != nil {
				return nil, err
			}

			d := &v12.Service{}
			if err := yaml.Unmarshal(buffer.Bytes(), d); err != nil {
				return nil, err
			}

			return d, nil
		},
		SetStatus: func(owner base.StepObject, target, now base.StepObject) (needUpdate bool, updateObject base.StepObject, err error) {
			c := owner.(*v14.Broker)
			o := now.(*v12.Service)
			c.Status.Dispatcher.Svc = o.Status

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
		Name: "dispatcher",
		Sub:  []*base.Step{svc, sts},
	}
}
