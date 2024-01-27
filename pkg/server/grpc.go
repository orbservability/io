package server

import (
	"net"
	"os"
	"time"

	"github.com/orbservability/telemetry/pkg/logs"
	"github.com/orbservability/telemetry/pkg/metrics"
	"github.com/orbservability/telemetry/pkg/traces"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// ServiceRegistrar is an interface that all gRPC services should implement.
type ServiceRegistrar interface {
	RegisterWithServer(*grpc.Server)
}

func ServeGRPC(errChan chan<- error, services []ServiceRegistrar) *grpc.Server {
	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(logs.UnaryServerInterceptor, metrics.UnaryServerInterceptor, traces.UnaryServerInterceptor),
		grpc.ChainStreamInterceptor(logs.StreamServerInterceptor, metrics.StreamServerInterceptor, traces.StreamServerInterceptor),
	)

	// Register each service with the gRPC server
	for _, service := range services {
		service.RegisterWithServer(grpcServer)
	}

	// Enable server reflection.
	// Server reflection is a feature that allows the server to describe its available services and methods.
	reflection.Register(grpcServer)

	// Create a listener on a specified port for the gRPC server
	portEnv := os.Getenv("PORT")
	if portEnv == "" {
		portEnv = "50051" // Default port number
	}
	lis, err := net.Listen("tcp", ":"+portEnv)
	if err != nil {
		errChan <- err
		return nil
	}

	// Start a gRPC server with the listener.
	log.Info().Msgf("Starting gRPC server on port %s", portEnv)
	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			errChan <- err
		}
	}()

	return grpcServer
}

func ShutdownGRPC(grpcServer *grpc.Server) {
	done := make(chan struct{})
	go func() {
		grpcServer.GracefulStop()
		close(done)
	}()

	select {
	case <-done:
		log.Info().Msg("gRPC server shut down gracefully")
	case <-time.After(shutdownTimeout()):
		grpcServer.Stop()
		log.Warn().Msg("gRPC server shutdown timed out")
	}
}
