package discovery

import (
	"fmt"
	v1 "github.com/stream-stack/store-operator/apis/knative/v1"
)

func GetPublisherName(b *v1.Broker) string {
	return fmt.Sprintf(`%s-publisher`, b.Name)
}

//
//func PublisherStoreSetPush(ctx context.Context, k8sClient client.Client, broker *v1.Broker) error {
//	var err error
//	list := &v15.StoreSetList{}
//	selectorMap, err := v13.LabelSelectorAsMap(broker.Spec.Selector)
//	if err != nil {
//		return err
//	}
//	err = k8sClient.List(ctx, list, client.MatchingLabels(selectorMap))
//	if err != nil {
//		return err
//	}
//	if len(list.Items) <= 0 {
//		return nil
//	}
//	data := buildStoreSetData(list.Items)
//
//	//推送storeset
//	addrs := buildPublisherPodAddr(broker)
//	action := func(addr string, client proto.XdsServiceClient) error {
//		_, err := client.StoreSetPush(ctx, &proto.StoreSetPushRequest{
//			Stores: data,
//		})
//		return err
//	}
//	c := make(chan error, 1)
//	for _, addr := range addrs {
//		ConnActionCh <- ConnAction{
//			Addr:   addr,
//			Action: action,
//			Result: c,
//		}
//	}
//	return <-c
//}
//
//func buildStoreSetData(items []v15.StoreSet) []*proto.StoreSet {
//	sets := make([]*proto.StoreSet, len(items))
//	for i, item := range items {
//		sets[i] = &proto.StoreSet{
//			Name:      item.Name,
//			Namespace: item.Namespace,
//			Uris:      buildStoreUri(item),
//		}
//	}
//	return sets
//}
//
//func DeletePublisherConn(broker *v1.Broker) {
//	addr := buildPublisherPodAddr(broker)
//	for _, s := range addr {
//		ConnDeleteCh <- ConnAction{
//			Addr: s,
//		}
//	}
//}
//
//func buildPublisherPodAddr(broker *v1.Broker) []string {
//	name := GetPublisherName(broker)
//	addrs := make([]string, broker.Spec.Publisher.Replicas)
//	for i := range addrs {
//		addrs[i] = fmt.Sprintf("%s-%d.%s.%s:%s", name, i, broker.Status.Publisher.SvcName, broker.Namespace, PublisherContainerPort)
//	}
//	return addrs
//}

const PublisherContainerPort = `8080`

func GetPublisherManagerContainerPort() string {
	return PublisherContainerPort
}
