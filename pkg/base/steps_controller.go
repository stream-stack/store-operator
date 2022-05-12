package base

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	configv1 "github.com/stream-stack/store-operator/apis/config/v1"
	v12 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"time"
)

var FinalizerName string
var retryDuration time.Duration

func DefaultConfigInit(config configv1.StreamControllerConfig) {
	FinalizerName = config.Workflow.FinalizerName
	retryDuration = config.Workflow.RetryDuration.Duration
}

type StepObject interface {
	runtime.Object
	metav1.Object
	metav1.ObjectMetaAccessor
}

type StepReconcile interface {
	GetClient() client.Client
	GetScheme() *runtime.Scheme
	GetRecorder() record.EventRecorder
}

type StatusDeepEqualFunc func(new StepObject, old runtime.Object) bool

type StepContext struct {
	context.Context
	StepObject
	logr.Logger
	StepReconcile
	deepEqualFunc StatusDeepEqualFunc
}

func NewStepContext(context context.Context, logger logr.Logger, r StepReconcile, deepEqualFunc StatusDeepEqualFunc) *StepContext {
	return &StepContext{Context: context, Logger: logger, StepReconcile: r, deepEqualFunc: deepEqualFunc}
}

func (c *StepContext) Reconcile(name types.NamespacedName, object StepObject, steps []*Step) (ctrl.Result, error) {
	err := c.GetClient().Get(c.Context, name, object)
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	c.StepObject = object
	if !object.GetObjectMeta().GetDeletionTimestamp().IsZero() {
		return c.deleteReconcile(steps)
	}
	return c.reconcile(steps, object.DeepCopyObject())
}

func (c *StepContext) deleteReconcile(steps []*Step) (ctrl.Result, error) {
	total := len(steps)
	for index, module := range steps {
		moduleName := module.Name
		moduleString := fmt.Sprintf("[%v/%v]%s ", index+1, total, moduleName)
		if err := module.Delete(c); err != nil {
			c.GetRecorder().Event(c.StepObject, v12.EventTypeWarning, "ModuleDeleteError", fmt.Sprintf("[%s] Error:%v", moduleName, err))
			c.Logger.Info(moduleString+"module delete error", "error", err)
			return ctrl.Result{}, err
		}
		c.Logger.Info(moduleString + "module delete success")
	}

	c.StepObject.SetFinalizers([]string{})
	c.Logger.Info("remove Finalizers...")
	if err := c.GetClient().Update(c.Context, c.StepObject); err != nil {
		c.Logger.Info("remove Finalizers error", "error", err)
		return ctrl.Result{}, err
	}
	c.Logger.Info("delete crd finish")

	return ctrl.Result{}, nil
}

func (c *StepContext) reconcile(steps []*Step, oldObject runtime.Object) (ctrl.Result, error) {
	total := len(steps)
	c.Logger.Info("Begin crd Reconcile", "totalStep:", total)
	for index, module := range steps {
		moduleName := module.Name
		moduleString := fmt.Sprintf("[%v/%v]%s ", index+1, total, moduleName)
		c.Logger.Info(moduleString)
		if err := module.Reconcile(c); err != nil {
			c.Logger.Info(moduleString+"module not exist,create error", "error", err)
			c.GetRecorder().Event(c.StepObject, v12.EventTypeWarning, "ReconcileError", fmt.Sprintf("[%s] Error:%v", moduleName, err))
			return ctrl.Result{
				Requeue:      true,
				RequeueAfter: retryDuration,
			}, err
		}
		ready, err := module.Ready(c)
		if err != nil {
			return ctrl.Result{
				Requeue:      true,
				RequeueAfter: retryDuration,
			}, err
		}
		if !ready {
			break
		}
	}
	if len(c.StepObject.GetFinalizers()) == 0 {
		c.StepObject.SetFinalizers([]string{FinalizerName})
		c.Logger.Info("update crd spec.finalizers...")
		if err := c.GetClient().Update(c.Context, c.StepObject); err != nil {
			c.Logger.Info("update crd spec.finalizers error", "error", err)
			return ctrl.Result{
				Requeue:      true,
				RequeueAfter: retryDuration,
			}, err
		}
	}
	if !c.deepEqualFunc(c.StepObject, oldObject) {
		c.Logger.Info("update crd status")
		if err := c.GetClient().Status().Update(c.Context, c.StepObject); err != nil {
			c.Logger.Info("update crd status error", "error", err)
			return ctrl.Result{
				Requeue:      true,
				RequeueAfter: retryDuration,
			}, err
		}
	}
	c.Logger.Info("End crd Reconcile")
	return ctrl.Result{}, nil
}

type Step struct {
	Name string
	Sub  []*Step

	GetObj func() StepObject
	Render func(c StepObject) (StepObject, error)
	//设置cluster的status,对比子资源目标和现在的声明情况
	SetStatus func(c StepObject, target, now StepObject) (needUpdate bool, updateObject StepObject, err error)
	Del       func(ctx context.Context, c StepObject, client client.Client) error
	Next      func(ctx *StepContext) (bool, error)

	SetDefault         func(c StepObject)
	ValidateCreateStep func(c StepObject) field.ErrorList
	ValidateUpdateStep func(now StepObject, old StepObject) field.ErrorList
}

func (m *Step) ValidateCreate(c StepObject) field.ErrorList {
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

func (m *Step) ValidateUpdate(now StepObject, old StepObject) field.ErrorList {
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

func (m *Step) Default(c StepObject) {
	if m.SetDefault != nil {
		m.SetDefault(c)
	}
	if m.hasSub() {
		for _, m := range m.Sub {
			m.Default(c)
		}
	}
}

func (m *Step) Reconcile(ctx *StepContext) error {
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

func (m *Step) exist(ctx *StepContext) (bool, error) {
	render, err := m.Render(ctx.StepObject)
	if err != nil {
		return false, err
	}
	err = ctx.StepReconcile.GetClient().Get(ctx, types.NamespacedName{
		Namespace: render.GetNamespace(),
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

func (m *Step) update(ctx *StepContext) error {
	obj := m.GetObj()
	render, err := m.Render(ctx.StepObject)
	if err != nil {
		return err
	}
	if err := ctx.StepReconcile.GetClient().Get(ctx, client.ObjectKey{
		Namespace: ctx.StepObject.GetNamespace(),
		Name:      render.GetName(),
	}, obj); err != nil {
		return err
	}

	needUpdate, o, err := m.SetStatus(ctx.StepObject, render, obj)
	if err != nil {
		return err
	}
	if needUpdate {
		if err := ctx.StepReconcile.GetClient().Update(ctx.Context, o); err != nil {
			return err
		}
	}
	return nil
}

func (m *Step) create(ctx *StepContext) error {
	render, err := m.Render(ctx.StepObject)
	if err != nil {
		return err
	}
	_ = controllerutil.SetControllerReference(ctx.StepObject, render, ctx.StepReconcile.GetScheme())
	ctx.StepReconcile.GetRecorder().Event(ctx.StepObject, v12.EventTypeNormal, fmt.Sprintf("Creating-%s", m.Name), render.GetName())
	err = ctx.StepReconcile.GetClient().Create(ctx, render)
	if err != nil && errors.IsAlreadyExists(err) {
		return nil
	}
	if err != nil {
		ctx.StepReconcile.GetRecorder().Event(ctx, v12.EventTypeWarning, "CreateError", fmt.Sprintf("%s,error:%v", render.GetName(), err))
	}
	return err
}

func (m *Step) Ready(ctx *StepContext) (bool, error) {
	if !m.hasSub() {
		if m.Next != nil {
			return m.Next(ctx)
		}
		return true, nil
	} else {
		for _, m := range m.Sub {
			ready, err := m.Ready(ctx)
			if !ready {
				return false, err
			}
		}
	}
	return true, nil
}

func (m *Step) Delete(ctx *StepContext) error {
	if !m.hasSub() {
		if m.Del != nil {
			err := m.Del(ctx, ctx.StepObject, ctx.StepReconcile.GetClient())
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
