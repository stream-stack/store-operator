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
var storesetlog = logf.Log.WithName("storeset-resource")

var StoreSetValidators = make([]base.ResourceValidator, 0)
var StoreSetDefaulters = make([]base.ResourceDefaulter, 0)

func (r *StoreSet) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// TODO(user): EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

//+kubebuilder:webhook:path=/mutate-core-stream-stack-tanx-v1-storeset,mutating=true,failurePolicy=fail,sideEffects=None,groups=core.stream-stack.tanx,resources=storesets,verbs=create;update,versions=v1,name=mstoreset.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &StoreSet{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *StoreSet) Default() {
	storesetlog.Info("default", "name", r.Name)

	// TODO(user): fill in your defaulting logic.
	for _, defaulter := range StoreSetDefaulters {
		defaulter.Default(r)
	}
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-core-stream-stack-tanx-v1-storeset,mutating=false,failurePolicy=fail,sideEffects=None,groups=core.stream-stack.tanx,resources=storesets,verbs=create;update,versions=v1,name=vstoreset.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &StoreSet{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *StoreSet) ValidateCreate() error {
	storesetlog.Info("validate create", "name", r.Name)

	var allErrs field.ErrorList

	for _, v := range StoreSetValidators {
		allErrs = append(allErrs, v.ValidateCreate(r)...)
	}

	if len(allErrs) == 0 {
		return nil
	}
	return errors.NewInvalid(r.TypeMeta.GroupVersionKind().GroupKind(), r.Name, allErrs)
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *StoreSet) ValidateUpdate(old runtime.Object) error {
	storesetlog.Info("validate update", "name", r.Name)

	oldC := old.(*StoreSet)
	var allErrs field.ErrorList

	for _, v := range StoreSetValidators {
		allErrs = append(allErrs, v.ValidateUpdate(oldC, r)...)
	}

	if len(allErrs) == 0 {
		return nil
	}
	return errors.NewInvalid(r.TypeMeta.GroupVersionKind().GroupKind(), r.Name, allErrs)
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *StoreSet) ValidateDelete() error {
	storesetlog.Info("validate delete", "name", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil
}
