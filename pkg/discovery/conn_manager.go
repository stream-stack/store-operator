package discovery

import (
	"context"
	grpc_retry "github.com/grpc-ecosystem/go-grpc-middleware/retry"
	"github.com/sirupsen/logrus"
	"github.com/stream-stack/store-operator/pkg/proto"
	"google.golang.org/grpc"
	"time"
)

var PushChan = make(chan PushAction, 1)
var DeleteConnChan = make(chan string, 1)

func StartPushChan(ctx context.Context) {
	defer func() {
		for _, conn := range connections {
			conn.Close()
		}
	}()
	for {
		select {
		case <-ctx.Done():
			return
		case action := <-PushChan:
			//获取grpc连接,如果没有,则新建
			conn, err := getConnection(action.Addr)
			if err != nil {
				logrus.Warnf("无法获取连接,%v", err)
				action.Result <- err
				continue
			}
			serviceClient := proto.NewXdsServiceClient(conn)
			err = action.Action(serviceClient)
			if err != nil {
				logrus.Warnf("推送出现错误,%v", err)
				action.Result <- err
			}
		case del := <-DeleteConnChan:
			logrus.Infof("清理连接%s", del)
			conn, ok := connections[del]
			if !ok {
				logrus.Infof("未找到连接%s,跳过", del)
				continue
			}
			delete(connections, del)
			_ = conn.Close()
			logrus.Infof("清理连接%s完成,连接关闭完成", del)
		}
	}
}

type PushAction struct {
	Addr   string
	Action func(client proto.XdsServiceClient) error
	Result chan error
}

var connections map[string]*grpc.ClientConn

func getConnection(addr string) (*grpc.ClientConn, error) {
	conn, ok := connections[addr]
	if !ok {
		conn, err := createConn(addr)
		if err != nil {
			return nil, err
		}
		connections[addr] = conn
		return conn, nil
	}
	return conn, nil
}

func createConn(addr string) (*grpc.ClientConn, error) {
	retryOpts := []grpc_retry.CallOption{
		grpc_retry.WithBackoff(grpc_retry.BackoffExponential(100 * time.Millisecond)),
		grpc_retry.WithMax(5),
	}
	return grpc.Dial(addr, grpc.WithInsecure(), grpc.WithUnaryInterceptor(grpc_retry.UnaryClientInterceptor(retryOpts...)))
}
