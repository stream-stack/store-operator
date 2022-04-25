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
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"time"
)

// log is for logging in this package.
var subscriptionlog = logf.Log.WithName("subscription-resource")

func (r *Subscription) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

//+kubebuilder:webhook:path=/mutate-knative-stream-stack-tanx-v1-subscription,mutating=true,failurePolicy=fail,sideEffects=None,groups=knative.stream-stack.tanx,resources=subscriptions,verbs=create;update,versions=v1,name=msubscription.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &Subscription{}

var DefaultMaxRetries = 3
var DefaultMaxRequestDuration = &metav1.Duration{Duration: 5 * time.Second}
var DefaultAckDuration = &metav1.Duration{Duration: 5 * time.Second}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *Subscription) Default() {
	subscriptionlog.Info("default", "name", r.Name)

	if r.Spec.Delivery == nil {
		r.Spec.Delivery = &SubscriberDelivery{
			MaxRetries:         &DefaultMaxRetries,
			MaxRequestDuration: DefaultMaxRequestDuration,
			AckDuration:        DefaultAckDuration,
		}
		return
	}
	if r.Spec.Delivery.MaxRequestDuration == nil {
		r.Spec.Delivery.MaxRequestDuration = DefaultMaxRequestDuration
	}
	if r.Spec.Delivery.MaxRetries == nil {
		r.Spec.Delivery.MaxRetries = &DefaultMaxRetries
	}
	if r.Spec.Delivery.AckDuration == nil {
		r.Spec.Delivery.AckDuration = DefaultAckDuration
	}

}

//+kubebuilder:webhook:path=/validate-knative-stream-stack-tanx-v1-subscription,mutating=false,failurePolicy=fail,sideEffects=None,groups=knative.stream-stack.tanx,resources=subscriptions,verbs=create;update,versions=v1,name=vsubscription.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &Subscription{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *Subscription) ValidateCreate() error {
	subscriptionlog.Info("validate create", "name", r.Name)
	var allErrs field.ErrorList
	if len(r.Spec.Broker) == 0 {
		allErrs = append(allErrs, field.Invalid(field.NewPath("spec.broker"), r.Spec.Broker, "broker is required"))
	}
	if r.Spec.Subscriber.Uri == nil && r.Spec.Subscriber.Service == nil && r.Spec.Subscriber.Pod == nil {
		allErrs = append(allErrs, field.Invalid(field.NewPath("spec.subscriber"), r.Spec.Subscriber, "subscriber is required"))
	}

	return errors.NewInvalid(r.TypeMeta.GroupVersionKind().GroupKind(), r.Name, allErrs)
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *Subscription) ValidateUpdate(old runtime.Object) error {
	subscriptionlog.Info("validate update", "name", r.Name)

	var allErrs field.ErrorList
	allErrs = append(allErrs, field.Invalid(field.NewPath("spec"), r.Spec, "spec is immutable"))

	return errors.NewInvalid(r.TypeMeta.GroupVersionKind().GroupKind(), r.Name, allErrs)
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *Subscription) ValidateDelete() error {
	subscriptionlog.Info("validate delete", "name", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil
}
