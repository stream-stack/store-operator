package storeset_arm

import (
	"github.com/stream-stack/store-operator/controllers"
	"github.com/stream-stack/store-operator/controllers/steps"
)

func init() {
	config := &controllers.InitConfig{
		Version:           "arm-1.0.0",
		StoreImage:        "",
		StoreReplicas:     3,
		PublisherImage:    "",
		PublisherReplicas: 1,
	}
	steps.NewLocalPersistentVolumeSteps(config)
	steps.NewStoreSteps(config)
}
