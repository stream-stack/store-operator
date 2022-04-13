package discovery

//
//import (
//	"context"
//	"fmt"
//	"github.com/sirupsen/logrus"
//	v1 "github.com/stream-stack/store-operator/apis/knative/v1"
//	"github.com/stream-stack/store-operator/pkg/proto"
//	"time"
//)
//
//type PartitionAllocator struct {
//	ctx           context.Context
//	cancelFunc    context.CancelFunc
//	serviceClient proto.XdsServiceClient
//	addr          string
//	group         *AllocatorGroupAdd
//	uri           string
//}
//
//func (a *PartitionAllocator) Stop() {
//	a.cancelFunc()
//}
//
//func (a *PartitionAllocator) Start(ctx context.Context) {
//	a.ctx, a.cancelFunc = context.WithCancel(ctx)
//	defer a.Stop()
//	for {
//		select {
//		case <-a.ctx.Done():
//			return
//		default:
//			partition, err := a.serviceClient.AllocatePartition(a.ctx, &proto.AllocatePartitionRequest{})
//			if err != nil {
//				logrus.Errorf("AllocatePartition error:%v , sleep 5 second for retry", err)
//				time.Sleep(time.Second * 5)
//			}
//			for {
//				select {
//				case <-a.ctx.Done():
//					return
//				default:
//					recv, err := partition.Recv()
//					if err != nil {
//						logrus.Errorf("AllocatePartition recv error:%v", err)
//						continue
//					}
//					a.group.AllocateCh <- recv
//				}
//			}
//		}
//	}
//}
//
//var Allocators = make(map[string]*PartitionAllocator)
//
//func getAllocator(broker v1.Broker) *PartitionAllocator {
//	name := getBrokerAllocatorName(broker)
//	return Allocators[name]
//}
//
//func getBrokerAllocatorName(broker v1.Broker) string {
//	return fmt.Sprintf("%s/%s", broker.Namespace, broker.Name)
//}
//
//func NewPartitionAllocator(group *AllocatorGroupAdd, uri string, serviceClient proto.XdsServiceClient) *PartitionAllocator {
//	return &PartitionAllocator{
//		group:         group,
//		serviceClient: serviceClient,
//		uri:           uri,
//	}
//}
