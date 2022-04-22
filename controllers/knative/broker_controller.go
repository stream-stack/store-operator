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

package knative

import (
	"context"
	knativev1 "github.com/stream-stack/store-operator/apis/knative/v1"
	"github.com/stream-stack/store-operator/pkg/base"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	"reflect"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// BrokerReconciler reconciles a Broker object
type BrokerReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

func (r *BrokerReconciler) GetClient() client.Client {
	return r.Client
}

func (r *BrokerReconciler) GetScheme() *runtime.Scheme {
	return r.Scheme
}

func (r *BrokerReconciler) GetRecorder() record.EventRecorder {
	return r.Recorder
}

//+kubebuilder:rbac:groups="",resources=serviceaccounts,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="rbac.authorization.k8s.io",resources=clusterrolebindings,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=knative.stream-stack.tanx,resources=brokers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=knative.stream-stack.tanx,resources=brokers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=knative.stream-stack.tanx,resources=brokers/finalizers,verbs=update
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps,resources=deployments/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=apps,resources=deployments/finalizers,verbs=update
//+kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=services/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=core,resources=services/finalizers,verbs=update
//+kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch;
//+kubebuilder:rbac:groups="",resources=events,verbs=create;patch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Broker object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.11.0/pkg/reconcile
func (r *BrokerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx)
	rl := l.WithValues("Broker", req.NamespacedName)

	return base.NewStepContext(ctx, rl, r, statusDeepEqual).Reconcile(req.NamespacedName, &knativev1.Broker{}, Steps)
}

// SetupWithManager sets up the controller with the Manager.
func (r *BrokerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&knativev1.Broker{}).Owns(&appsv1.Deployment{}).Owns(&corev1.Service{}).Owns(&corev1.ServiceAccount{}).
		Complete(r)
}

func statusDeepEqual(now base.StepObject, old runtime.Object) bool {
	nowSet := now.(*knativev1.Broker)
	oldSet := old.(*knativev1.Broker)
	return reflect.DeepEqual(nowSet.Status, oldSet.Status)
}

var Steps = make([]*base.Step, 0)
