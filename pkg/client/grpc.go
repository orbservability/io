package client

import (
	"github.com/orbservability/telemetry/pkg/logs"
	"google.golang.org/grpc"
)

// ClientRegistrar is an interface that all gRPC clients should implement.
type ClientRegistrar interface {
	RegisterClient(conn *grpc.ClientConn)
}

func DialGRPC(targetURL string, client ClientRegistrar, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	defaultOpts := []grpc.DialOption{
		grpc.WithChainUnaryInterceptor(logs.UnaryClientInterceptor),
		grpc.WithChainStreamInterceptor(logs.StreamClientInterceptor),
	}

	conn, err := grpc.Dial(targetURL, append(defaultOpts, opts...)...)
	if err != nil {
		return nil, err
	}

	client.RegisterClient(conn)

	return conn, nil
}
