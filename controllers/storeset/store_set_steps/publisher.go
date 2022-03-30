package store_set_steps

import (
	_ "embed"
	"fmt"
	v12 "github.com/stream-stack/store-operator/apis/storeset/v1"
	"github.com/stream-stack/store-operator/pkg/base"
	"github.com/stream-stack/store-operator/pkg/store_client"
	v1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/util/yaml"
	"reflect"
)

//go:embed publisher_template.yaml
var deptTemplate []byte

func NewPublisherSteps(cfg *InitConfig) *base.Step {
	var publisher = &base.Step{
		Name: "publisherDept",
		GetObj: func() base.StepObject {
			return &v1.Deployment{}
		},
		Render: func(t base.StepObject) (base.StepObject, error) {
			c := t.(*v12.StoreSet)
			d := &v1.Deployment{}
			err := yaml.Unmarshal(deptTemplate, d)
			if err != nil {
				return nil, err
			}
			d.Name = c.Name
			d.Namespace = c.Namespace
			d.Labels = c.Labels
			d.Spec.Selector.MatchLabels = c.Labels
			d.Spec.Template.Labels = c.Labels
			d.Spec.Replicas = c.Spec.Publisher.Replicas
			container := d.Spec.Template.Spec.Containers[0]
			container.Image = c.Spec.Publisher.Image
			container.Args = []string{fmt.Sprintf(container.Args[0], c.Status.StoreStatus.ServiceName, c.Namespace, store_client.StoreContainerPort.IntVal)}
			d.Spec.Template.Spec.Containers[0] = container

			return d, nil
		},
		SetStatus: func(owner base.StepObject, target, now base.StepObject) (needUpdate bool, updateObject base.StepObject, err error) {
			c := owner.(*v12.StoreSet)
			o := now.(*v1.Deployment)
			c.Status.PublisherStatus.Name = c.Name
			c.Status.PublisherStatus.Status = o.Status

			t := target.(*v1.Deployment)
			if !reflect.DeepEqual(t.Spec, o.Spec) {
				o.Spec = t.Spec
				return true, o, nil
			}

			return false, now, nil
		},
		Next: func(ctx *base.StepContext) (bool, error) {
			c := ctx.StepObject.(*v12.StoreSet)
			return c.Status.PublisherStatus.Status.AvailableReplicas == *c.Spec.Publisher.Replicas, nil
		},
		SetDefault: func(t base.StepObject) {
			c := t.(*v12.StoreSet)
			if c.Spec.Publisher.Image == "" {
				c.Spec.Publisher.Image = cfg.PublisherImage
			}
			if c.Spec.Publisher.Replicas == nil {
				c.Spec.Publisher.Replicas = &cfg.PublisherReplicas
			}
		},
	}

	var step = &base.Step{
		Name: "publisher",
		Sub:  []*base.Step{publisher},
	}
	return step
}
