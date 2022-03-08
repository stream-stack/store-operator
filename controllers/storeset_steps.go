package controllers

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	v13 "github.com/stream-stack/store-operator/api/v1"
	v12 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const FinalizerName = "finalizer.storeset.stream-stack.tanx"

var VersionsModules = make(map[string][]*Step)

type Object interface {
	runtime.Object
	metav1.Object
	metav1.ObjectMetaAccessor
}

func AddSteps(key string, m ...*Step) {
	value, ok := VersionsModules[key]
	if !ok {
		value = make([]*Step, 0)
	}
	value = append(value, m...)

	for i := 0; i < len(value); i++ {
		for j := i; j < len(value); j++ {
			if value[i].Order > value[j].Order {
				value[i], value[j] = value[j], value[i]
			}
		}
	}

	VersionsModules[key] = value
	for _, module := range m {
		v13.RegisterVersionedDefaulters(key, module)
		v13.RegisterVersionedValidators(key, module)
	}
}

type ModuleContext struct {
	context.Context
	*v13.StoreSet
	logr.Logger
	reconciler *StoreSetReconciler
	Old        *v13.StoreSet
}

func NewStepContext(context context.Context, c *v13.StoreSet, logger logr.Logger, r *StoreSetReconciler) *ModuleContext {
	return &ModuleContext{Context: context, StoreSet: c, Logger: logger, reconciler: r, Old: c.DeepCopy()}
}

type InitConfig struct {
	Version           string
	StoreImage        string
	StoreReplicas     int32
	PublisherImage    string
	PublisherReplicas int32
}

type Step struct {
	Name  string
	Sub   []*Step
	Order int

	GetObj func() Object
	Render func(c *v13.StoreSet) Object
	//设置cluster的status,对比子资源目标和现在的声明情况
	SetStatus func(c *v13.StoreSet, target, now Object) (needUpdate bool, updateObject Object, err error)
	Del       func(ctx context.Context, c *v13.StoreSet, client client.Client) error
	Next      func(c *v13.StoreSet) bool

	SetDefault         func(c *v13.StoreSet)
	ValidateCreateStep func(c *v13.StoreSet) field.ErrorList
	ValidateUpdateStep func(now *v13.StoreSet, old *v13.StoreSet) field.ErrorList
}

func (m *Step) ValidateCreate(c *v13.StoreSet) field.ErrorList {
	if !m.hasSub() {
		if m.ValidateCreateStep == nil {
			return nil
		}
		return m.ValidateCreateStep(c)
	} else {
		for _, m := range m.Sub {
			if err := m.ValidateCreate(c); err != nil {
				return err
			}
		}
	}
	return nil
}

func (m *Step) ValidateUpdate(now *v13.StoreSet, old *v13.StoreSet) field.ErrorList {
	if !m.hasSub() {
		if m.ValidateUpdateStep == nil {
			return nil
		}
		return m.ValidateUpdateStep(now, old)
	} else {
		for _, m := range m.Sub {
			if err := m.ValidateUpdate(now, old); err != nil {
				return err
			}
		}
	}
	return nil
}

func (m *Step) Default(c *v13.StoreSet) {
	if m.SetDefault != nil {
		m.SetDefault(c)
	}
	if m.hasSub() {
		for _, m := range m.Sub {
			m.Default(c)
		}
	}
}

func (m *Step) Reconcile(ctx *ModuleContext) error {
	if !m.hasSub() {
		exist, err := m.exist(ctx)
		if err != nil {
			return err
		}
		if exist {
			if err := m.update(ctx); err != nil {
				return err
			}
		} else {
			if err := m.create(ctx); err != nil {
				return err
			}
		}
	} else {
		for _, m := range m.Sub {
			if err := m.Reconcile(ctx); err != nil {
				return err
			}
		}
	}
	return nil
}

func (m *Step) hasSub() bool {
	if m.Sub == nil || len(m.Sub) == 0 {
		return false
	}
	return true
}

func (m *Step) exist(ctx *ModuleContext) (bool, error) {
	render := m.Render(ctx.StoreSet)
	err := ctx.reconciler.Get(ctx, types.NamespacedName{
		Namespace: ctx.Namespace,
		Name:      render.GetName(),
	}, m.GetObj())
	if err != nil && errors.IsNotFound(err) {
		return false, nil
	}
	if err == nil {
		return true, nil
	}
	return false, err
}

func (m *Step) update(ctx *ModuleContext) error {
	obj := m.GetObj()
	render := m.Render(ctx.StoreSet)
	if err := ctx.reconciler.Get(ctx, client.ObjectKey{
		Namespace: ctx.Namespace,
		Name:      render.GetName(),
	}, obj); err != nil {
		return err
	}

	needUpdate, o, err := m.SetStatus(ctx.StoreSet, render, obj)
	if err != nil {
		return err
	}
	if needUpdate {
		if err := ctx.reconciler.Update(ctx.Context, o); err != nil {
			return err
		}
	}
	return nil
}

func (m *Step) create(ctx *ModuleContext) error {
	render := m.Render(ctx.StoreSet)
	if err := controllerutil.SetControllerReference(ctx.StoreSet, render, ctx.reconciler.Scheme); err != nil {
		return err
	}
	ctx.reconciler.Recorder.Event(ctx.StoreSet, v12.EventTypeNormal, fmt.Sprintf("Creating-%s", m.Name), render.GetName())
	err := ctx.reconciler.Create(ctx, render)
	if err != nil && errors.IsAlreadyExists(err) {
		return nil
	}
	if err != nil {
		ctx.reconciler.Recorder.Event(ctx, v12.EventTypeWarning, "CreateError", fmt.Sprintf("%s,error:%v", render.GetName(), err))
	}
	return err
}

func (m *Step) Ready(ctx *ModuleContext) bool {
	if !m.hasSub() {
		if m.Next != nil {
			return m.Next(ctx.StoreSet)
		}
		return true
	} else {
		for _, m := range m.Sub {
			if !m.Ready(ctx) {
				return false
			}
		}
	}
	return true
}

func (m *Step) Delete(ctx *ModuleContext) error {
	if !m.hasSub() {
		if m.Del != nil {
			err := m.Del(ctx, ctx.StoreSet, ctx.reconciler.Client)
			if err != nil {
				ctx.Logger.Error(err, "invoke module Del error")
			}
			return err
		}
	}
	for _, m := range m.Sub {
		if err := m.Delete(ctx); err != nil {
			return err
		}
	}
	return nil
}
