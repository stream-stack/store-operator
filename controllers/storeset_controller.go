/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"fmt"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/record"
	"reflect"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	v1 "github.com/stream-stack/store-operator/api/v1"
)

// StoreSetReconciler reconciles a StoreSet object
type StoreSetReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

//+kubebuilder:rbac:groups=core.stream-stack.tanx,resources=storesets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core.stream-stack.tanx,resources=storesets/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=core.stream-stack.tanx,resources=storesets/finalizers,verbs=update
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps,resources=deployments/status,verbs=get
//+kubebuilder:rbac:groups=apps,resources=deployments/finalizers,verbs=update
//+kubebuilder:rbac:groups=apps,resources=statefulset,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps,resources=statefulset/status,verbs=get
//+kubebuilder:rbac:groups=apps,resources=statefulset/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch
//+kubebuilder:rbac:groups="",resources=configmaps/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=persistentvolume,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=persistentvolume/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=service,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=service/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the StoreSet closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the StoreSet object against the actual StoreSet state, and then
// perform operations to make the StoreSet state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.11.0/pkg/reconcile
func (r *StoreSetReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx)
	rl := l.WithValues("StoreSet", req.NamespacedName)

	cs := &v1.StoreSet{}
	err := r.Client.Get(ctx, req.NamespacedName, cs)
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	moduleContext := NewStepContext(ctx, cs, rl, r)
	if !cs.ObjectMeta.DeletionTimestamp.IsZero() {
		return deleteStoreSet(moduleContext)
	}

	return ReconcileStoreSet(moduleContext)
}

func deleteStoreSet(ctx *ModuleContext) (ctrl.Result, error) {
	version := ctx.Spec.Version
	modules, ok := VersionsModules[version]
	if !ok {
		return ctrl.Result{}, fmt.Errorf("not support version %s", version)
	}
	total := len(modules)
	for index, module := range modules {
		moduleName := module.Name
		moduleString := fmt.Sprintf("[%v/%v]%s ", index+1, total, moduleName)
		if err := module.Delete(ctx); err != nil {
			ctx.reconciler.Recorder.Event(ctx, corev1.EventTypeWarning, "ModuleDeleteError", fmt.Sprintf("[%s] Error:%v", moduleName, err))
			ctx.Info(moduleString+"module delete error", "error", err)
			return ctrl.Result{}, err
		}
		ctx.Info(moduleString + "module delete success")
	}

	ctx.SetFinalizers([]string{})
	ctx.Info("remove Finalizers...")
	if err := ctx.reconciler.Update(ctx, ctx.StoreSet); err != nil {
		ctx.Info("remove Finalizers error", "error", err)
		return ctrl.Result{}, err
	}
	ctx.Info("delete StoreSet finish", "name", ctx.Name, "namespace", ctx.Namespace)

	return ctrl.Result{}, nil
}

var retryDuration = time.Second * 1

func ReconcileStoreSet(ctx *ModuleContext) (ctrl.Result, error) {
	version := ctx.Spec.Version
	modules, ok := VersionsModules[version]
	if !ok {
		return ctrl.Result{Requeue: false}, fmt.Errorf("not support version %s", version)
	}
	ctx.Info("Begin StoreSet Reconcile", "version", version, "name", ctx.Name, "namespace", ctx.Namespace)
	total := len(modules)
	for index, module := range modules {
		moduleName := module.Name
		moduleString := fmt.Sprintf("[%v/%v]%s ", index+1, total, moduleName)
		ctx.Info(moduleString, "version", version, "name", ctx.Name, "namespace", ctx.Namespace)
		if err := module.Reconcile(ctx); err != nil {
			ctx.Info(moduleString+"module not exist,create error", "error", err)
			ctx.reconciler.Recorder.Event(ctx, corev1.EventTypeWarning, "ReconcileError", fmt.Sprintf("[%s] Error:%v", moduleName, err))
			return ctrl.Result{
				Requeue:      true,
				RequeueAfter: retryDuration,
			}, err
		}
		if !module.Ready(ctx) {
			break
		}
	}
	ctx.Info("update StoreSet(crd) status")
	if len(ctx.GetFinalizers()) == 0 {
		ctx.SetFinalizers([]string{FinalizerName})
	}
	if !reflect.DeepEqual(ctx.Old.Status, ctx.StoreSet.Status) {
		if err := ctx.reconciler.Client.Status().Update(ctx, ctx.StoreSet); err != nil {
			ctx.Info("update StoreSet(crd) status error", "error", err)
			return ctrl.Result{
				Requeue:      true,
				RequeueAfter: retryDuration,
			}, err
		}
	}
	ctx.Info("End StoreSet Reconcile", "version", version)

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *StoreSetReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1.StoreSet{}).Owns(&appsv1.Deployment{}).Owns(&appsv1.StatefulSet{}).Owns(&corev1.Service{}).
		Owns(&corev1.ConfigMap{}).Owns(&corev1.PersistentVolume{}).
		Complete(r)
}
