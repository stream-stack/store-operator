package discovery

//
//import (
//	"context"
//	grpc_retry "github.com/grpc-ecosystem/go-grpc-middleware/retry"
//	"github.com/sirupsen/logrus"
//	"github.com/stream-stack/store-operator/pkg/proto"
//	"google.golang.org/grpc"
//	"time"
//)
//
//var ConnActionCh = make(chan ConnAction, 1)
//var ConnDeleteCh = make(chan ConnAction, 1)
//
////type PartitionAllocator struct {
////	Addr   string
////	Action func(client proto.XdsServiceClient) error
////	Result chan error
////}
////
////var PartitionAllocatorAddCh = make(chan PartitionAllocator, 1)
////var PartitionAllocatorDeleteCh = make(chan PartitionAllocator, 1)
//
//func StartPushChan(ctx context.Context) {
//	defer func() {
//		for _, conn := range connections {
//			conn.Close()
//		}
//	}()
//	for {
//		select {
//		case <-ctx.Done():
//			return
//		case action := <-ConnActionCh:
//			//获取grpc连接,如果没有,则新建
//			conn, err := getConnection(action.Addr)
//			if err != nil {
//				logrus.Warnf("无法获取连接,%v", err)
//				action.Result <- err
//				close(action.Result)
//				continue
//			}
//			if action.Action != nil {
//				serviceClient := proto.NewXdsServiceClient(conn)
//				err = action.Action(action.Addr, serviceClient)
//				if err != nil {
//					logrus.Warnf("推送出现错误,%v", err)
//					action.Result <- err
//				}
//			}
//			close(action.Result)
//		case del := <-ConnDeleteCh:
//			logrus.Infof("清理连接%s", del.Addr)
//			if del.Action != nil {
//				_ = del.Action(del.Addr, nil)
//			}
//			conn, ok := connections[del.Addr]
//			if !ok {
//				logrus.Infof("未找到连接%s,跳过", del.Addr)
//				continue
//			}
//			delete(connections, del.Addr)
//			_ = conn.Close()
//			logrus.Infof("清理连接%s完成,连接关闭完成", del.Addr)
//		}
//	}
//}
//
//type ConnAction struct {
//	Addr   string
//	Action func(addr string, client proto.XdsServiceClient) error
//	Result chan error
//}
//
//var connections map[string]*grpc.ClientConn
//
//func getConnection(addr string) (*grpc.ClientConn, error) {
//	conn, ok := connections[addr]
//	if !ok {
//		conn, err := createConn(addr)
//		if err != nil {
//			return nil, err
//		}
//		connections[addr] = conn
//		return conn, nil
//	}
//	return conn, nil
//}
//
//func createConn(addr string) (*grpc.ClientConn, error) {
//	retryOpts := []grpc_retry.CallOption{
//		grpc_retry.WithBackoff(grpc_retry.BackoffExponential(100 * time.Millisecond)),
//		grpc_retry.WithMax(5),
//	}
//	return grpc.Dial(addr, grpc.WithInsecure(), grpc.WithUnaryInterceptor(grpc_retry.UnaryClientInterceptor(retryOpts...)))
//}
