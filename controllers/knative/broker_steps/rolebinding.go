package broker_steps

import (
	"bytes"
	"context"
	"embed"
	_ "embed"
	"fmt"
	configv1 "github.com/stream-stack/store-operator/apis/config/v1"
	v14 "github.com/stream-stack/store-operator/apis/knative/v1"
	"github.com/stream-stack/store-operator/pkg/base"
	v12 "k8s.io/api/core/v1"
	v13 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/util/yaml"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"text/template"
)

//go:embed broker_serviceaccount_template.yaml broker_clusterrolebinding_template.yaml
var roleBindingTemplateFs embed.FS

var roleBindingYamlTemplate *template.Template

func init() {
	var err error
	roleBindingYamlTemplate, err = template.New("rolebinding").Funcs(map[string]interface{}{
		//"GetBrokerServiceAccountName": discovery.GetBrokerServiceAccountName,
		//"GetClusterRoleName":          discovery.GetClusterRoleName,
	}).ParseFS(roleBindingTemplateFs, "*")
	if err != nil {
		panic(err)
	}
}

func NewRoleBinding(cfg configv1.StreamControllerConfig) *base.Step {
	account := &base.Step{
		Name: fmt.Sprintf(`rolebinding-serviceaccount`),
		GetObj: func() base.StepObject {
			return &v12.ServiceAccount{}
		},
		Render: func(t base.StepObject) (base.StepObject, error) {
			c := t.(*v14.Broker)
			buffer := &bytes.Buffer{}
			err := roleBindingYamlTemplate.ExecuteTemplate(buffer, "broker_serviceaccount_template.yaml", c)
			if err != nil {
				return nil, err
			}

			d := &v12.ServiceAccount{}
			err = yaml.Unmarshal(buffer.Bytes(), d)
			if err != nil {
				return nil, err
			}

			return d, nil
		},
		SetStatus: func(owner base.StepObject, target, now base.StepObject) (needUpdate bool, updateObject base.StepObject, err error) {
			return false, nil, nil
		},
	}
	binding := &base.Step{
		Name: fmt.Sprintf(`rolebinding-binding`),
		GetObj: func() base.StepObject {
			return &v13.ClusterRoleBinding{}
		},
		Render: func(t base.StepObject) (base.StepObject, error) {
			c := t.(*v14.Broker)
			buffer := &bytes.Buffer{}
			if err := roleBindingYamlTemplate.ExecuteTemplate(buffer, "broker_clusterrolebinding_template.yaml", c); err != nil {
				return nil, err
			}

			d := &v13.ClusterRoleBinding{}
			if err := yaml.Unmarshal(buffer.Bytes(), d); err != nil {
				return nil, err
			}

			return d, nil
		},
		SetStatus: func(owner base.StepObject, target, now base.StepObject) (needUpdate bool, updateObject base.StepObject, err error) {
			c := owner.(*v14.Broker)
			c.Status.Uuid = c.Spec.Uuid
			o := now.(*v13.ClusterRoleBinding)

			t := target.(*v13.ClusterRoleBinding)
			if !reflect.DeepEqual(t.RoleRef, o.RoleRef) {
				o.RoleRef = t.RoleRef
				return true, o, nil
			}
			if !reflect.DeepEqual(t.Subjects, o.Subjects) {
				o.Subjects = t.Subjects
				return true, o, nil
			}
			return false, now, nil
		},
		Next: func(ctx *base.StepContext) (bool, error) {
			return true, nil
		},
		Del: func(ctx context.Context, t base.StepObject, client client.Client) error {
			c := t.(*v14.Broker)
			buffer := &bytes.Buffer{}
			if err := roleBindingYamlTemplate.ExecuteTemplate(buffer, "broker_clusterrolebinding_template.yaml", c); err != nil {
				return err
			}

			d := &v13.ClusterRoleBinding{}
			if err := yaml.Unmarshal(buffer.Bytes(), d); err != nil {
				return err
			}
			return client.Delete(ctx, d)
		},
	}
	return &base.Step{
		Name: "rolebinding",
		Sub:  []*base.Step{account, binding},
	}
}
