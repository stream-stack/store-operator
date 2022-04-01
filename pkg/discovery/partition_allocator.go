package discovery

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	v1 "github.com/stream-stack/store-operator/apis/knative/v1"
	v15 "github.com/stream-stack/store-operator/apis/storeset/v1"
	"github.com/stream-stack/store-operator/pkg/proto"
	"github.com/stream-stack/store-operator/pkg/store_client"
	v13 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"math/rand"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type PartitionAllocator struct {
	broker        v1.Broker
	client        client.Client
	ctx           context.Context
	cancelFunc    context.CancelFunc
	serviceClient proto.XdsServiceClient
	addr          string
}

func (a *PartitionAllocator) Stop() {
	a.cancelFunc()
	//name := getBrokerAllocatorName(a.broker)
	//ConnDeleteCh <- ConnAction{
	//	Addr: a.addr,
	//	Action: func(addr string, client proto.XdsServiceClient) error {
	//		delete(Allocators, name)
	//		return nil
	//	},
	//}
}

func (a *PartitionAllocator) Start(ctx context.Context) {
	cancel, cancelFunc := context.WithCancel(ctx)
	a.ctx = cancel
	a.cancelFunc = cancelFunc
	defer a.Stop()
	for {
		select {
		case <-a.ctx.Done():
			return
		default:
			partition, err := a.serviceClient.AllocatePartition(cancel, &proto.AllocatePartitionRequest{})
			if err != nil {
				panic(err)
			}
			recv, err := partition.Recv()
			if err != nil {
				panic(err)
			}
			fmt.Println(recv)
			//TODO: 收到分片分配请求后,计算分片,并分配

			list := &v15.StoreSetList{}
			selectorMap, err := v13.LabelSelectorAsMap(a.broker.Spec.Selector)
			if err != nil {
				panic(err)
			}
			err = a.client.List(ctx, list, client.MatchingLabels(selectorMap))
			if err != nil {
				panic(err)
			}
			if len(list.Items) <= 0 {
				continue
			}
			storesetData := buildStoreData(list)

			//TODO:根据分片规则分片
			intn := rand.Intn(len(storesetData))
			set := storesetData[intn]
			marshal, err := json.Marshal(proto.Partition{
				RangeRegexp: "[0-9]{1,5}",
				Store: proto.Store{
					Name:      set.Name,
					Namespace: set.Namespace,
					Uris:      set.Uris,
				},
			})
			if err != nil {
				logrus.Errorf("序列化分片数据错误,%v", err)
				panic(err)
			}
			apply, err := store_client.Apply(a.ctx, set.Uris, &proto.ApplyRequest{
				StreamName: systemBrokerPartition,
				StreamId:   GetStreamName(&a.broker),
				EventId:    "1",
				Data:       marshal,
			})
			if err != nil {
				logrus.Errorf("写入分片出现错误,%v", err)
			}
			logrus.Debugf("写入分片完成,返回值:%+v", apply)
		}
	}
}

var Allocators = make(map[string]*PartitionAllocator)

func getAllocator(broker v1.Broker) *PartitionAllocator {
	name := getBrokerAllocatorName(broker)
	return Allocators[name]
}

func getBrokerAllocatorName(broker v1.Broker) string {
	return fmt.Sprintf("%s/%s", broker.Namespace, broker.Name)
}

func NewPartitionAllocator(broker v1.Broker, k8sClient client.Client, serviceClient proto.XdsServiceClient, addr string) *PartitionAllocator {
	return &PartitionAllocator{
		broker:        broker,
		client:        k8sClient,
		serviceClient: serviceClient,
		addr:          addr,
	}
}
