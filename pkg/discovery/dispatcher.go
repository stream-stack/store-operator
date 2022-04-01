package discovery

import (
	"context"
	"encoding/json"
	"fmt"
	_ "github.com/Jille/grpc-multi-resolver"
	"github.com/sirupsen/logrus"
	v1 "github.com/stream-stack/store-operator/apis/knative/v1"
	v15 "github.com/stream-stack/store-operator/apis/storeset/v1"
	"github.com/stream-stack/store-operator/pkg/proto"
	"github.com/stream-stack/store-operator/pkg/store_client"
	_ "google.golang.org/grpc/health"
	v13 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/rand"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"time"
)

func DispatcherStoreSetPush(ctx context.Context, k8sClient client.Client, broker *v1.Broker) error {
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
	storesetData := buildStoreData(list)
	//生成sts地址
	adds := buildDispatcherAddress(broker)
	b := *broker
	action := func(addr string, client proto.XdsServiceClient) error {
		//启动分片器
		allocator := NewPartitionAllocator(b, k8sClient, client, addr)
		Allocators[getBrokerAllocatorName(b)] = allocator
		go allocator.Start(ctx)

		_, err := client.StoreSetPush(ctx, &proto.StoreSetPushRequest{
			Stores: storesetData,
		})
		return err
	}
	c := make(chan error, 1)
	for _, addr := range adds {
		ConnActionCh <- ConnAction{
			Addr:   addr,
			Action: action,
			Result: c,
		}
	}
	return <-c
}

//func StartPartitionAllocator(ctx context.Context, k8sClient client.Client, broker *v1.Broker) error {
//	//生成sts地址
//	adds := buildDispatcherAddress(broker)
//	b := *broker
//	action := func(addr string, client proto.XdsServiceClient) error {
//		//1.检查是否存在Allocator
//		allocator := getAllocator(b)
//		if allocator != nil {
//			return nil
//		}
//		//2.如果不存在,则启动
//		allocator = NewPartitionAllocator(b, k8sClient, client, addr)
//		Allocators[getBrokerAllocatorName(b)] = allocator
//		go allocator.Start(ctx)
//		return nil
//	}
//	c := make(chan error, 1)
//	for _, addr := range adds {
//		ConnActionCh <- ConnAction{
//			Addr:   addr,
//			Action: action,
//			Result: c,
//		}
//	}
//	return <-c
//}
//
//func StopPartitionAllocator(broker *v1.Broker) {
//	//生成sts地址
//	adds := buildDispatcherAddress(broker)
//	action := func(addr string, client proto.XdsServiceClient) error {
//		allocator := getAllocator(*broker)
//		if allocator == nil {
//			return nil
//		}
//		allocator.Stop()
//		return nil
//	}
//	for _, addr := range adds {
//		ConnActionCh <- ConnAction{
//			Addr:   addr,
//			Action: action,
//		}
//	}
//}

const systemBrokerPartition = "_system_broker_partition"

func allocatePartition(ctx context.Context, items []v15.StoreSet, data []proto.Store, broker *v1.Broker) error {
	logrus.Debugf("开始分配分片,storeset:%d", len(items))
	timeout, cancelFunc := context.WithTimeout(ctx, time.Second*10)
	defer cancelFunc()
	//TODO:根据分片规则分片
	intn := rand.Intn(len(items))
	//set := items[intn]
	s := data[intn]
	//buffer := &bytes.Buffer{}
	marshal, err := json.Marshal(proto.Partition{
		RangeRegexp: "[0-9]{1,5}",
		Store:       s,
	})
	if err != nil {
		logrus.Errorf("序列化分片数据错误,%v", err)
		return err
	}
	apply, err := store_client.Apply(timeout, s.Uris, &proto.ApplyRequest{
		StreamName: systemBrokerPartition,
		StreamId:   GetStreamName(broker),
		EventId:    "1",
		Data:       marshal,
	})
	if err != nil {
		logrus.Errorf("写入分片出现错误,%v", err)
		return err
	}
	logrus.Debugf("写入分片完成,返回值:%+v", apply)
	return nil
}

func GetStreamName(b *v1.Broker) string {
	return fmt.Sprintf("%s-%s", b.Namespace, b.Name)
}

func GetDispatcherStsName(b *v1.Broker) string {
	return fmt.Sprintf(`%s-dispatcher`, b.Name)
}

func buildStoreData(list *v15.StoreSetList) []*proto.StoreSet {
	data := make([]*proto.StoreSet, len(list.Items))
	for i, item := range list.Items {
		data[i] = &proto.StoreSet{Uris: buildStoreUri(item), Name: item.Name, Namespace: item.Namespace}
	}
	return data
}

func buildStoreUri(item v15.StoreSet) []string {
	replicas := *item.Spec.Store.Replicas
	addrs := make([]string, replicas)
	var i int32
	for ; i < replicas; i++ {
		//TODO:重构名称的生成,应该和模板统一,使用template的自定义函数
		addrs[i] = fmt.Sprintf(`%s-%d.%s.%s:%s`, item.Name, i, item.Status.StoreStatus.ServiceName, item.Namespace, store_client.StoreContainerPort.String())
	}
	return addrs
}

func buildDispatcherAddress(broker *v1.Broker) []string {
	replicas := broker.Spec.Dispatcher.Replicas
	addrs := make([]string, replicas)
	var i int32
	for ; i < replicas; i++ {
		addrs[i] = fmt.Sprintf(`%s-%d.%s.%s:%s`, GetDispatcherStsName(broker), i, broker.Status.Dispatcher.SvcName, broker.GetNamespace(), DispatcherContainerPort)
	}
	return addrs
}

func DeleteDispatcherConn(broker *v1.Broker) {
	address := buildDispatcherAddress(broker)
	for _, s := range address {
		ConnDeleteCh <- ConnAction{
			Addr: s,
			Action: func(addr string, client proto.XdsServiceClient) error {
				allocator := getAllocator(*broker)
				if allocator == nil {
					return nil
				}
				allocator.Stop()
				name := getBrokerAllocatorName(*broker)
				delete(Allocators, name)
				return nil
			},
		}
	}
}

const DispatcherContainerPort = `8080`

func GetDispatcherContainerPort() string {
	return DispatcherContainerPort
}
