package steps

import (
	_ "embed"
	"fmt"
	corev1 "github.com/stream-stack/store-operator/api/v1"
	"github.com/stream-stack/store-operator/controllers"
	"gopkg.in/yaml.v3"
	v1 "k8s.io/api/apps/v1"
	"reflect"
)

//go:embed publisher_template.yaml
var deptTemplate []byte

func NewPublisherSteps(cfg *controllers.InitConfig) {
	var publisher = &controllers.Step{
		Name: "publisherDept",
		GetObj: func() controllers.Object {
			return &v1.Deployment{}
		},
		Render: func(c *corev1.StoreSet) controllers.Object {
			d := &v1.Deployment{}
			_ = yaml.Unmarshal(deptTemplate, d)
			d.Name = c.Name
			d.Namespace = c.Namespace
			d.Labels = c.Labels
			d.Spec.Selector.MatchLabels = c.Labels
			d.Spec.Template.Labels = c.Labels
			d.Spec.Replicas = c.Spec.Publisher.Replicas
			container := d.Spec.Template.Spec.Containers[0]
			container.Image = c.Spec.Publisher.Image
			container.Args = []string{fmt.Sprintf(container.Args[0], c.Status.StoreStatus.ServiceName, c.Namespace, containerPort.IntVal)}
			d.Spec.Template.Spec.Containers[0] = container

			return d
		},
		SetStatus: func(c *corev1.StoreSet, target, now controllers.Object) (needUpdate bool, updateObject controllers.Object, err error) {
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
		Next: func(ctx *controllers.ModuleContext) (bool, error) {
			c := ctx.StoreSet
			return c.Status.PublisherStatus.Status.AvailableReplicas == *c.Spec.Publisher.Replicas, nil
		},
		SetDefault: func(c *corev1.StoreSet) {
			if c.Spec.Publisher.Image == "" {
				c.Spec.Publisher.Image = cfg.PublisherImage
			}
			if c.Spec.Publisher.Replicas == nil {
				c.Spec.Publisher.Replicas = &cfg.PublisherReplicas
			}
		},
	}

	var step = &controllers.Step{
		Order: 30,
		Name:  "publisher",
		Sub:   []*controllers.Step{publisher},
	}
	controllers.AddSteps(cfg.Version, step)
}
