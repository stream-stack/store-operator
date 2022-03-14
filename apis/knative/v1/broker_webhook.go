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

package v1

import (
	"github.com/stream-stack/store-operator/pkg/base"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var brokerlog = logf.Log.WithName("broker-resource")

var BrokerValidators = make([]base.ResourceValidator, 0)
var BrokerDefaulters = make([]base.ResourceDefaulter, 0)

func (r *Broker) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

//+kubebuilder:webhook:path=/mutate-knative-stream-stack-tanx-v1-broker,mutating=true,failurePolicy=fail,sideEffects=None,groups=knative.stream-stack.tanx,resources=brokers,verbs=create;update;delete,versions=v1,name=mbroker.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &Broker{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *Broker) Default() {
	brokerlog.Info("default", "name", r.Name)

	for _, defaulter := range BrokerDefaulters {
		defaulter.Default(r)
	}
}

//+kubebuilder:webhook:path=/validate-knative-stream-stack-tanx-v1-broker,mutating=false,failurePolicy=fail,sideEffects=None,groups=knative.stream-stack.tanx,resources=brokers,verbs=create;update;delete,versions=v1,name=vbroker.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &Broker{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *Broker) ValidateCreate() error {
	brokerlog.Info("validate create", "name", r.Name)

	var allErrs field.ErrorList

	for _, v := range BrokerValidators {
		allErrs = append(allErrs, v.ValidateCreate(r)...)
	}

	if len(allErrs) == 0 {
		return nil
	}
	return errors.NewInvalid(r.TypeMeta.GroupVersionKind().GroupKind(), r.Name, allErrs)
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *Broker) ValidateUpdate(old runtime.Object) error {
	brokerlog.Info("validate update", "name", r.Name)
	oldC := old.(*Broker)
	var allErrs field.ErrorList

	for _, v := range BrokerValidators {
		allErrs = append(allErrs, v.ValidateUpdate(oldC, r)...)
	}

	if len(allErrs) == 0 {
		return nil
	}
	return errors.NewInvalid(r.TypeMeta.GroupVersionKind().GroupKind(), r.Name, allErrs)
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *Broker) ValidateDelete() error {
	brokerlog.Info("validate delete", "name", r.Name)

	return nil
}
