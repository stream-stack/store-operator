package store_set_steps

import (
	"fmt"
	configv1 "github.com/stream-stack/store-operator/apis/config/v1"
	v13 "github.com/stream-stack/store-operator/apis/storeset/v1"
	"github.com/stream-stack/store-operator/pkg/base"
	v12 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const AppLabelKey = `app`

func buildAppLabelName(name string) string {
	return fmt.Sprintf(`%s-storeset`, name)
}

//TODO:修改为使用yaml+go template渲染资源(更直观)

func NewLocalPersistentVolumeSteps(cfg configv1.StreamControllerConfig) *base.Step {
	replicas := cfg.Store.Replicas
	steps := make([]*base.Step, replicas)
	var i int32 = 0
	for ; i < replicas; i++ {
		steps[i] = buildStepByIndex(i, cfg)
	}
	return &base.Step{
		Name: "localPersistentVolume",
		Sub:  steps,
		SetDefault: func(t base.StepObject) {
			c := t.(*v13.StoreSet)
			//全局label处理,加入app:xxx(name)-storeset,version:arm-1.0.0标签
			_, ok := c.Labels[AppLabelKey]
			if !ok {
				c.Labels[AppLabelKey] = buildAppLabelName(c.Name)
			}
		},
	}
}

func buildStepByIndex(index int32, cfg configv1.StreamControllerConfig) *base.Step {
	return &base.Step{
		Name: fmt.Sprintf(`localPersistentVolume-%d`, index),
		GetObj: func() base.StepObject {
			return &v12.PersistentVolume{}
		},
		Render: func(t base.StepObject) (base.StepObject, error) {
			c := t.(*v13.StoreSet)
			filesystem := v12.PersistentVolumeFilesystem
			return &v12.PersistentVolume{
				ObjectMeta: metav1.ObjectMeta{
					Name:        fmt.Sprintf(`%s-%d`, c.Name, index),
					Labels:      c.Labels,
					Annotations: c.Annotations,
				},
				Spec: v12.PersistentVolumeSpec{
					Capacity: v12.ResourceList{
						v12.ResourceStorage: c.Spec.Volume.Capacity,
					},
					PersistentVolumeSource: v12.PersistentVolumeSource{
						Local: c.Spec.Volume.LocalVolumeSource,
					},
					AccessModes:                   []v12.PersistentVolumeAccessMode{v12.ReadWriteOnce},
					PersistentVolumeReclaimPolicy: v12.PersistentVolumeReclaimRetain,
					StorageClassName:              c.Name,
					VolumeMode:                    &filesystem,
					NodeAffinity:                  c.Spec.Volume.NodeAffinity,
				},
			}, nil
		},
		SetStatus: func(owner base.StepObject, target, now base.StepObject) (needUpdate bool, updateObject base.StepObject, err error) {
			c := owner.(*v13.StoreSet)
			o := now.(*v12.PersistentVolume)
			c.Status.VolumeStatus = v13.LocalPvStatus{
				Name:   c.Name,
				Status: o.Status,
			}

			t := target.(*v12.PersistentVolume)
			if !reflect.DeepEqual(t.Spec, o.Spec) {
				o.Spec = t.Spec
				return true, o, nil
			}
			c.Status.Status = v13.StoreSetStatusPVCreating

			return false, now, nil
		},
		Next: func(ctx *base.StepContext) (bool, error) {
			//note: 命名空间中的资源(StoreSet)无法owner pv,cluster-scoped resource must not have a namespace-scoped owner, owner's namespace default
			//只能判断是否创建成功, 如果创建成功, 则认为通过
			p := &v12.PersistentVolume{}
			p.Name = fmt.Sprintf(`%s-%d`, ctx.StepObject.GetName(), index)
			if err := ctx.GetClient().Get(ctx.Context, client.ObjectKeyFromObject(p), p); err != nil {
				ctx.Logger.Error(err, "get PersistentVolume error", "PersistentVolume.Name", p.Name)
				return false, err
			}
			return true, nil
			//return p.Status.Phase == v12.VolumeAvailable || p.Status.Phase == v12.VolumeBound
		},
		SetDefault: func(t base.StepObject) {
			c := t.(*v13.StoreSet)
			if c.Spec.Volume.LocalVolumeSource == nil {
				c.Spec.Volume.LocalVolumeSource = &v12.LocalVolumeSource{
					Path: cfg.Store.VolumePath,
				}
			}
		},
	}
}
