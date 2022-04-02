package discovery

import (
	"context"
	pp "github.com/golang/protobuf/proto"
	"github.com/sirupsen/logrus"
	"github.com/stream-stack/store-operator/apis/knative/v1"
	v15 "github.com/stream-stack/store-operator/apis/storeset/v1"
	"github.com/stream-stack/store-operator/pkg/proto"
	"github.com/stream-stack/store-operator/pkg/store_client"
	v13 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"math/rand"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type AllocatorGroupAdd struct {
	broker     v1.Broker
	k8sClient  client.Client
	uris       []string
	brokerName string
	resultCh   chan error
	AllocateCh chan *proto.AllocatePartitionResponse
	cancelFunc context.CancelFunc
	ctx        context.Context
}

func (g *AllocatorGroupAdd) Stop() {
	g.cancelFunc()
}

func (g *AllocatorGroupAdd) Start(ctx context.Context) {
	defer func() {
		close(g.resultCh)
	}()
	g.AllocateCh = make(chan *proto.AllocatePartitionResponse, 1)
	g.ctx, g.cancelFunc = context.WithCancel(ctx)

	action := func(addr string, client proto.XdsServiceClient) error {
		//启动分片器
		allocator := NewPartitionAllocator(g, addr, client)
		Allocators[getBrokerAllocatorName(g.broker)] = allocator
		go allocator.Start(g.ctx)
		return nil
	}
	for _, addr := range g.uris {
		c := make(chan error, 1)
		ConnActionCh <- ConnAction{
			Addr:   addr,
			Action: action,
			Result: c,
		}
	}
	//TODO: 如何处理返回?
	go func() {
		var partitionCount uint64
		for {
			select {
			case <-ctx.Done():
				return
			case allocate := <-g.AllocateCh:
				if partitionCount > allocate.PartitionCount {
					continue
				}
				success := g.allocatePartition(allocate)
				if success {
					partitionCount = allocate.PartitionCount
				}
			}
		}
	}()

}

func (g *AllocatorGroupAdd) getStoreSetData() ([]*proto.StoreSet, error) {
	list := &v15.StoreSetList{}
	selectorMap, err := v13.LabelSelectorAsMap(g.broker.Spec.Selector)
	if err != nil {
		logrus.Errorf("LabelSelectorAsMap error,%v", err)
		return nil, err
	}
	err = g.k8sClient.List(g.ctx, list, client.MatchingLabels(selectorMap))
	if err != nil {
		logrus.Errorf("k8s client list storeset error,%v", err)
		return nil, err
	}
	if len(list.Items) <= 0 {
		return make([]*proto.StoreSet, 0), nil
	}
	return buildStoreData(list), nil
}

func (g *AllocatorGroupAdd) allocatePartition(allocate *proto.AllocatePartitionResponse) bool {
	storesetData, err := g.getStoreSetData()
	if err != nil {
		return false
	}

	//TODO:根据分片规则分片,分片规则应该写在旧store和新store中
	intn := rand.Intn(len(storesetData))
	set := storesetData[intn]

	bytes, err := pp.Marshal(&proto.Partition{
		Begin: 0,
		Store: set,
	})
	if err != nil {
		logrus.Errorf("protobuf marshal partition error,%v", err)
		return false
	}
	apply, err := store_client.Apply(g.ctx, set.Uris, &proto.ApplyRequest{
		StreamName: systemBrokerPartition,
		StreamId:   GetStreamName(&g.broker),
		EventId:    1,
		Data:       bytes,
	})
	if err != nil {
		logrus.Errorf("write partition error,%v", err)
		return false
	}
	logrus.Debugf("write partition success,result:%+v", apply)
	return true
}

var AllocatorGroups = make(map[string]*AllocatorGroupAdd)
var AllocatorGroupAddCh = make(chan AllocatorGroupAdd, 1)
var AllocatorGroupDelCh = make(chan string, 1)

func StartAllocatorGroup(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case add := <-AllocatorGroupAddCh:
			group, ok := AllocatorGroups[add.brokerName]
			if ok {
				continue
			}
			group = &add
			AllocatorGroups[add.brokerName] = group
			group.Start(ctx)
		case name := <-AllocatorGroupDelCh:
			group, ok := AllocatorGroups[name]
			if !ok {
				continue
			}
			delete(AllocatorGroups, name)
			group.Stop()
		}
	}
}
