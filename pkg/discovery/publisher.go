package discovery

import (
	"context"
	"fmt"
	v1 "github.com/stream-stack/store-operator/apis/knative/v1"
	v15 "github.com/stream-stack/store-operator/apis/storeset/v1"
	"github.com/stream-stack/store-operator/pkg/proto"
	v13 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func GetPublisherStsName(b *v1.Broker) string {
	return fmt.Sprintf(`%s-publisher`, b.Name)
}

func PublisherStoreSetPush(ctx context.Context, k8sClient client.Client, broker *v1.Broker) error {
	var err error
	list := &v15.StoreSetList{}
	selectorMap, err := v13.LabelSelectorAsMap(broker.Spec.Selector)
	if err != nil {
		return err
	}
	err = k8sClient.List(ctx, list, client.MatchingLabels(selectorMap))
	if err != nil {
		return err
	}
	if len(list.Items) <= 0 {
		return nil
	}
	data := buildStoreSetData(list.Items)

	//推送storeset
	addrs := buildPublisherPodAddr(broker)
	action := func(client proto.XdsServiceClient) error {
		_, err := client.StoreSetPush(ctx, &proto.StoreSetPushRequest{
			Stores: data,
		})
		return err
	}
	for _, addr := range addrs {
		PushChan <- PushAction{
			Addr:   addr,
			Action: action,
		}
	}
	//TODO:这里如何传出推送错误等情况?
	return nil
}

func buildStoreSetData(items []v15.StoreSet) []*proto.StoreSet {
	sets := make([]*proto.StoreSet, len(items))
	for i, item := range items {
		sets[i] = &proto.StoreSet{
			Name:      item.Name,
			Namespace: item.Namespace,
			Uris:      buildStoreUri(item),
		}
	}
	return sets
}

func buildPublisherPodAddr(broker *v1.Broker) []string {
	name := GetPublisherStsName(broker)
	addrs := make([]string, broker.Spec.Publisher.Replicas)
	for i := range addrs {
		addrs[i] = fmt.Sprintf("%s-%d.%s.%s:%s", name, i, broker.Status.Publisher.SvcName, broker.Namespace, PublisherContainerPort)
	}
	return addrs
}

const PublisherContainerPort = `8080`

func GetPublisherContainerPort() string {
	return PublisherContainerPort
}
