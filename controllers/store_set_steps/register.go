package store_set_steps

import (
	v1 "github.com/stream-stack/store-operator/api/v1"
	"github.com/stream-stack/store-operator/controllers"
)

func init() {
	config := &InitConfig{
		StoreImage:        "ccr.ccs.tencentyun.com/stream/store:latest",
		StoreReplicas:     3,
		PublisherImage:    "ccr.ccs.tencentyun.com/stream/publisher:latest",
		PublisherReplicas: 1,
	}
	pv := NewLocalPersistentVolumeSteps(config)
	v1.StoreSetValidators = append(v1.StoreSetValidators, pv)
	v1.StoreSetDefaulters = append(v1.StoreSetDefaulters, pv)
	controllers.Steps = append(controllers.Steps, pv)
	sts := NewStoreSteps(config)
	v1.StoreSetValidators = append(v1.StoreSetValidators, sts)
	v1.StoreSetDefaulters = append(v1.StoreSetDefaulters, sts)
	controllers.Steps = append(controllers.Steps, sts)
	publisher := NewPublisherSteps(config)
	v1.StoreSetValidators = append(v1.StoreSetValidators, publisher)
	v1.StoreSetDefaulters = append(v1.StoreSetDefaulters, publisher)
	controllers.Steps = append(controllers.Steps, publisher)
}

type InitConfig struct {
	StoreImage        string
	StoreReplicas     int32
	PublisherImage    string
	PublisherReplicas int32
}
