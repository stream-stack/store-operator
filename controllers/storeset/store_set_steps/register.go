package store_set_steps

import (
	configv1 "github.com/stream-stack/store-operator/apis/config/v1"
	v12 "github.com/stream-stack/store-operator/apis/storeset/v1"
	"github.com/stream-stack/store-operator/controllers/storeset"
)

func Register(config configv1.StreamControllerConfig) {
	pv := NewLocalPersistentVolumeSteps(config)
	v12.StoreSetValidators = append(v12.StoreSetValidators, pv)
	v12.StoreSetDefaulters = append(v12.StoreSetDefaulters, pv)
	storeset.Steps = append(storeset.Steps, pv)
	sts := NewStoreSteps(config)
	v12.StoreSetValidators = append(v12.StoreSetValidators, sts)
	v12.StoreSetDefaulters = append(v12.StoreSetDefaulters, sts)
	storeset.Steps = append(storeset.Steps, sts)
}
