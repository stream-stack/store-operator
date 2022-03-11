package store_set_steps

import (
	v12 "github.com/stream-stack/store-operator/apis/storeset/v1"
	"github.com/stream-stack/store-operator/controllers/storeset"
)

func init() {
	config := &InitConfig{
		StoreImage:        "ccr.ccs.tencentyun.com/stream/store:latest",
		StoreReplicas:     3,
		PublisherImage:    "ccr.ccs.tencentyun.com/stream/publisher:latest",
		PublisherReplicas: 1,
	}
	pv := NewLocalPersistentVolumeSteps(config)
	v12.StoreSetValidators = append(v12.StoreSetValidators, pv)
	v12.StoreSetDefaulters = append(v12.StoreSetDefaulters, pv)
	storeset.Steps = append(storeset.Steps, pv)
	sts := NewStoreSteps(config)
	v12.StoreSetValidators = append(v12.StoreSetValidators, sts)
	v12.StoreSetDefaulters = append(v12.StoreSetDefaulters, sts)
	storeset.Steps = append(storeset.Steps, sts)
	publisher := NewPublisherSteps(config)
	v12.StoreSetValidators = append(v12.StoreSetValidators, publisher)
	v12.StoreSetDefaulters = append(v12.StoreSetDefaulters, publisher)
	storeset.Steps = append(storeset.Steps, publisher)
}

type InitConfig struct {
	StoreImage        string
	StoreReplicas     int32
	PublisherImage    string
	PublisherReplicas int32
}
