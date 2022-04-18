package discovery

import (
	"fmt"
	_ "github.com/Jille/grpc-multi-resolver"
	"github.com/sirupsen/logrus"
	v1 "github.com/stream-stack/store-operator/apis/knative/v1"
	_ "google.golang.org/grpc/health"
	"k8s.io/apimachinery/pkg/util/json"
)

//
//func StartAllocatorGroupWithBroker(ctx context.Context, k8sClient client.Client, broker *v1.Broker) error {
//	//生成sts地址
//	adds := buildDispatcherAddress(broker)
//	b := *broker
//
//	resultCh := make(chan error, 1)
//	AllocatorGroupAddCh <- AllocatorGroupAdd{
//		brokerName: GetStreamName(broker),
//		broker:     b,
//		k8sClient:  k8sClient,
//		uris:       adds,
//		resultCh:   resultCh,
//	}
//	return <-resultCh
//}
//
//func DispatcherStoreSetPush(ctx context.Context, k8sClient client.Client, broker *v1.Broker) error {
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
//	storesetData := buildStoreData(list)
//	//生成sts地址
//	adds := buildDispatcherAddress(broker)
//	action := func(addr string, client proto.XdsServiceClient) error {
//		_, err := client.StoreSetPush(ctx, &proto.StoreSetPushRequest{
//			Stores: storesetData,
//		})
//		return err
//	}
//	for _, addr := range adds {
//		c := make(chan error, 1)
//		ConnActionCh <- ConnAction{
//			Addr:   addr,
//			Action: action,
//			Result: c,
//		}
//
//		if err := <-c; err != nil {
//			return err
//		}
//	}
//	return nil
//}

const systemBrokerPartition = "_system_broker_partition"

func GetStreamName(b *v1.Broker) string {
	return fmt.Sprintf("%s-%s-%s", b.Namespace, b.Name, b.Status.Uuid)
}
func GetSelector(b *v1.Broker) string {
	marshal, err := json.Marshal(b.Spec.Selector)
	if err != nil {
		logrus.Warnf("marshal selector error:%v", err)
	}
	return string(marshal)
}

func GetDispatcherDeptName(b *v1.Broker) string {
	return fmt.Sprintf(`%s-dispatcher`, b.Name)
}

//
//func buildStoreData(list *v15.StoreSetList) []*proto.StoreSet {
//	data := make([]*proto.StoreSet, len(list.Items))
//	for i, item := range list.Items {
//		data[i] = &proto.StoreSet{Uris: buildStoreUri(item), Name: item.Name, Namespace: item.Namespace}
//	}
//	return data
//}
//
//func buildStoreUri(item v15.StoreSet) []string {
//	replicas := *item.Spec.Store.Replicas
//	addrs := make([]string, replicas)
//	var i int32
//	for ; i < replicas; i++ {
//		//TODO:重构名称的生成,应该和模板统一,使用template的自定义函数
//		addrs[i] = fmt.Sprintf(`%s-%d.%s.%s:%s`, item.Name, i, item.Status.StoreStatus.ServiceName, item.Namespace, store_client.StoreContainerPort.String())
//	}
//	return addrs
//}
//
//func buildDispatcherAddress(broker *v1.Broker) []string {
//	replicas := broker.Spec.Dispatcher.Replicas
//	addrs := make([]string, replicas)
//	var i int32
//	for ; i < replicas; i++ {
//		addrs[i] = fmt.Sprintf(`%s-%d.%s.%s:%s`, GetDispatcherDeptName(broker), i, broker.Status.Dispatcher.SvcName, broker.GetNamespace(), DispatcherManagerContainerPort)
//	}
//	return addrs
//}
//
//func DeleteDispatcherConn(broker *v1.Broker) {
//	address := buildDispatcherAddress(broker)
//	for _, s := range address {
//		ConnDeleteCh <- ConnAction{
//			Addr: s,
//			Action: func(addr string, client proto.XdsServiceClient) error {
//				allocator := getAllocator(*broker)
//				if allocator == nil {
//					return nil
//				}
//				allocator.Stop()
//				name := getBrokerAllocatorName(*broker)
//				delete(Allocators, name)
//				return nil
//			},
//		}
//	}
//}
//func DeleteAllocator(broker *v1.Broker) {
//	AllocatorGroupDelCh <- GetStreamName(broker)
//}

const DispatcherManagerContainerPort = `8080`

func GetDispatcherManagerContainerPort() string {
	return DispatcherManagerContainerPort
}
