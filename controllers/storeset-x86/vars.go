package storeset_x86

import (
	"github.com/stream-stack/store-operator/controllers"
	"github.com/stream-stack/store-operator/controllers/steps"
)

func init() {
	config := &controllers.InitConfig{
		Version:           "x86-1.0.0",
		StoreImage:        "ccr.ccs.tencentyun.com/stream/store:latest",
		StoreReplicas:     3,
		PublisherImage:    "ccr.ccs.tencentyun.com/stream/publisher:latest",
		PublisherReplicas: 1,
	}
	steps.NewLocalPersistentVolumeSteps(config)
	steps.NewStoreSteps(config)
	steps.NewPublisherSteps(config)
}
