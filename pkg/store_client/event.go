package store_client

import (
	"context"
	_ "github.com/Jille/grpc-multi-resolver"
	grpc_retry "github.com/grpc-ecosystem/go-grpc-middleware/retry"
	protocol "github.com/stream-stack/store-operator/pkg/proto"
	"google.golang.org/grpc"
	_ "google.golang.org/grpc/health"
	"time"
)

func Apply(ctx context.Context, urls string, request *protocol.ApplyRequest) (*protocol.ApplyResponse, error) {
	serviceConfig := `{"healthCheckConfig": {"serviceName": "store"}, "loadBalancingConfig": [ { "round_robin": {} } ]}`
	retryOpts := []grpc_retry.CallOption{
		grpc_retry.WithBackoff(grpc_retry.BackoffExponential(100 * time.Millisecond)),
		grpc_retry.WithMax(5),
	}
	conn, err := grpc.Dial(urls,
		grpc.WithDefaultServiceConfig(serviceConfig), grpc.WithInsecure(),
		grpc.WithDefaultCallOptions(grpc.WaitForReady(true)),
		grpc.WithUnaryInterceptor(grpc_retry.UnaryClientInterceptor(retryOpts...)))
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	eventCli := protocol.NewEventServiceClient(conn)
	return eventCli.Apply(ctx, request)
}
