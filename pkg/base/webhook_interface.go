package base

import "k8s.io/apimachinery/pkg/util/validation/field"

type ResourceValidator interface {
	ValidateCreate(c StepObject) field.ErrorList
	ValidateUpdate(now StepObject, old StepObject) field.ErrorList
}

type ResourceDefaulter interface {
	Default(c StepObject)
}
