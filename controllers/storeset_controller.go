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
	_ "embed"
	"github.com/stream-stack/store-operator/pkg/base"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	"reflect"
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

func (r *StoreSetReconciler) GetClient() client.Client {
	return r.Client
}

func (r *StoreSetReconciler) GetScheme() *runtime.Scheme {
	return r.Scheme
}

func (r *StoreSetReconciler) GetRecorder() record.EventRecorder {
	return r.Recorder
}

//+kubebuilder:rbac:groups=core.stream-stack.tanx,resources=storesets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core.stream-stack.tanx,resources=storesets/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=core.stream-stack.tanx,resources=storesets/finalizers,verbs=update
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps,resources=deployments/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=apps,resources=deployments/finalizers,verbs=update
//+kubebuilder:rbac:groups=apps,resources=statefulsets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps,resources=statefulsets/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=apps,resources=statefulsets/finalizers,verbs=update
//+kubebuilder:rbac:groups=core,resources=configmaps,verbs=get;list;watch
//+kubebuilder:rbac:groups=core,resources=configmaps/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=core,resources=configmaps/finalizers,verbs=update
//+kubebuilder:rbac:groups=core,resources=persistentvolumes,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=services/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=core,resources=services/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=events,verbs=create;patch

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

	//TODO:整理ctx参数至Reconcile方法中
	return base.NewStepContext(ctx, rl, r, statusDeepEqual).Reconcile(req.NamespacedName, &v1.StoreSet{}, Steps)
}

// SetupWithManager sets up the controller with the Manager.
func (r *StoreSetReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1.StoreSet{}).Owns(&appsv1.Deployment{}).Owns(&appsv1.StatefulSet{}).Owns(&corev1.Service{}).
		Owns(&corev1.ConfigMap{}).Owns(&corev1.PersistentVolume{}).
		Complete(r)
}

func statusDeepEqual(now base.StepObject, old runtime.Object) bool {
	nowSet := now.(*v1.StoreSet)
	oldSet := old.(*v1.StoreSet)
	return reflect.DeepEqual(nowSet.Status, oldSet.Status)
}

var Steps = make([]*base.Step, 0)
