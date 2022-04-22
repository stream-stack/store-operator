package store_client

import (
	"context"
	_ "github.com/Jille/grpc-multi-resolver"
	grpc_retry "github.com/grpc-ecosystem/go-grpc-middleware/retry"
	"github.com/sirupsen/logrus"
	"github.com/stream-stack/common/protocol/store"
	"google.golang.org/grpc"
	_ "google.golang.org/grpc/health"
	"k8s.io/apimachinery/pkg/util/intstr"
	"strings"
	"time"
)

var StoreContainerPort = intstr.FromInt(50051)

func Apply(ctx context.Context, urls []string, request *store.ApplyRequest) (*store.ApplyResponse, error) {
	targetAddress := storeAddressFormat(urls)
	serviceConfig := `{"healthCheckConfig": {"serviceName": "store"}, "loadBalancingConfig": [ { "round_robin": {} } ]}`
	retryOpts := []grpc_retry.CallOption{
		grpc_retry.WithBackoff(grpc_retry.BackoffExponential(100 * time.Millisecond)),
		grpc_retry.WithMax(5),
	}
	logrus.Debugf("apply地址为:%s", urls)
	conn, err := grpc.Dial(targetAddress,
		grpc.WithDefaultServiceConfig(serviceConfig), grpc.WithInsecure(),
		grpc.WithDefaultCallOptions(grpc.WaitForReady(true)),
		grpc.WithUnaryInterceptor(grpc_retry.UnaryClientInterceptor(retryOpts...)))
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	eventCli := store.NewEventServiceClient(conn)
	return eventCli.Apply(ctx, request)
}

func storeAddressFormat(urls []string) string {
	return "multi:///" + strings.Join(urls, ",")
}
