package steps

import (
	"fmt"
	v1 "github.com/stream-stack/store-operator/api/v1"
	"github.com/stream-stack/store-operator/controllers"
	v12 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const DefaultVolumePath = "/data"
const AppLabelKey = `app`
const VersionLabelKey = `version`

func buildAppLabelName(name string) string {
	return fmt.Sprintf(`%s-storeset`, name)
}

func NewLocalPersistentVolumeSteps(cfg *controllers.InitConfig) {
	steps := make([]*controllers.Step, cfg.StoreReplicas)
	var i int32 = 0
	for ; i < cfg.StoreReplicas; i++ {
		steps[i] = buildStepByIndex(i)
	}

	controllers.AddSteps(cfg.Version, &controllers.Step{
		Order: 10,
		Name:  "localPersistentVolume",
		Sub:   steps,
		SetDefault: func(c *v1.StoreSet) {
			//全局label处理,加入app:xxx(name)-storeset,version:arm-1.0.0标签
			_, ok := c.Labels[AppLabelKey]
			if !ok {
				c.Labels[AppLabelKey] = buildAppLabelName(c.Name)
			}
			_, ok = c.Labels[VersionLabelKey]
			if !ok {
				c.Labels[VersionLabelKey] = c.Spec.Version
			}
		},
	})
}

func buildStepByIndex(index int32) *controllers.Step {
	return &controllers.Step{
		Name: fmt.Sprintf(`localPersistentVolume-%d`, index),
		GetObj: func() controllers.Object {
			return &v12.PersistentVolume{}
		},
		Render: func(c *v1.StoreSet) controllers.Object {
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
			}
		},
		SetStatus: func(c *v1.StoreSet, target, now controllers.Object) (needUpdate bool, updateObject controllers.Object, err error) {
			o := now.(*v12.PersistentVolume)
			c.Status.VolumeStatus = v1.LocalPvStatus{
				Name:   c.Name,
				Status: o.Status,
			}

			t := target.(*v12.PersistentVolume)
			if !reflect.DeepEqual(t.Spec, o.Spec) {
				o.Spec = t.Spec
				return true, o, nil
			}

			return false, now, nil
		},
		Next: func(ctx *controllers.ModuleContext) (bool, error) {
			//note: 命名空间中的资源(StoreSet)无法owner pv,cluster-scoped resource must not have a namespace-scoped owner, owner's namespace default
			//只能判断是否创建成功, 如果创建成功, 则认为通过
			p := &v12.PersistentVolume{}
			p.Name = fmt.Sprintf(`%s-%d`, ctx.StoreSet.Name, index)
			if err := ctx.GetK8sClient().Get(ctx.Context, client.ObjectKeyFromObject(p), p); err != nil {
				ctx.Logger.Error(err, "get PersistentVolume error", "PersistentVolume.Name", p.Name)
				return false, err
			}
			return true, nil
			//return p.Status.Phase == v12.VolumeAvailable || p.Status.Phase == v12.VolumeBound
		},
		SetDefault: func(c *v1.StoreSet) {
			if c.Spec.Volume.LocalVolumeSource == nil {
				c.Spec.Volume.LocalVolumeSource = &v12.LocalVolumeSource{
					Path: DefaultVolumePath,
				}
			}
		},
	}
}
