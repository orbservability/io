# I/O

Common setup for incoming and outgoing requests. E.g. Client and server setup.

## Installation

```go
require (
  github.com/orbservability/io v0.0.3
)
```

## Usage

### Input

Examples

```go
package thing

import (
  "fmt"

  pb "github.com/orbservability/schemas/v1"
  "google.golang.org/grpc"
  "google.golang.org/protobuf/proto"
)

type ServiceHandler struct {
  pb.UnimplementedThingServiceServer
}

func NewServiceHandler() (*ServiceHandler, error) {
  return &ServiceHandler{}, nil
}

func (s *ServiceHandler) RegisterWithServer(grpcServer *grpc.Server) {
  pb.RegisterThingServiceServer(grpcServer, s)
}

func (s *ServiceHandler) DoThings(stream pb.ThingService_DoThingsServer) error {
  for {
    message, err := stream.Recv()
    if err != nil {
      return err
    }

    // Do the thing
    fmt.Println(proto.MessageName(message))
  }
}
```

```go
package main

import (
  "net/http"
  "os"
  "os/signal"
  "syscall"

  "github.com/prometheus/client_golang/prometheus/promhttp"
  "github.com/rs/zerolog/log"
  "orbservability.com/myapp/pkg/thing"
  "github.com/orbservability/io/pkg/server"
)

func main() {
  errChan := make(chan error, 2) // Error channel for server errors

  // Map HTTP routes to handlers, and serve HTTP
  handlers := map[string]http.Handler{
    "/metrics": promhttp.Handler(),
  }
  httpServer := server.ServeHTTP(errChan, handlers)

  // Initialize gRPC services, and serve gRPC
  service, err := thing.NewServiceHandler()
  if err != nil {
    log.Fatal().Err(err).Msg("Error initializing gRPC service")
  }
  services := []server.ServiceRegistrar{
    service,
  }
  grpcServer := server.ServeGRPC(errChan, services)

  // Set up signal handling for graceful shutdown, and wait for a termination signal
  sigChan := make(chan os.Signal, 1)
  signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

  // Block until a signal is received
  select {
  case <-sigChan:
    log.Warn().Msg("Application shutting down")

    server.ShutdownGRPC(grpcServer)
    server.ShutdownHTTP(httpServer)
  case err := <-errChan:
    log.Fatal().Err(err).Msg("Server error")
  }
}
```

### Output

Examples

```go
import (
  "context"

  "github.com/orbservability/io/pkg/kafka"
  "github.com/orbservability/io/pkg/schema"
  "github.com/orbservability/io/pkg/client"
  "github.com/rs/zerolog/log"
  "github.com/twmb/franz-go/pkg/kgo"

  pb "github.com/orbservability/schemas/v1"
)

func main() {
  ctx := context.Background()

  // Work with Kafka
  kafkaClient, err := kafka.NewClient()
  if err != nil {
    log.Fatal().Err(err).Msg("Error initializing Kafka client")
  }
  defer kafkaClient.Close()

  // Initialize gRPC client
  thing := &thing.ServiceClient{}
  grpcConn, err := client.DialGRPC("api.orbservability.com", thing, grpc.WithTransportCredentials(insecure.NewCredentials()))
  if err != nil {
    log.Fatal().Err(err).Msg("Error creating gRPC connection")
  }
  defer grpcConn.Close()
  grpcStream, err := thing.DoThings(ctx)
  if err != nil {
    log.Fatal().Err(err).Msg("Error creating gRPC stream")
  }
  defer grpcStream.CloseAndRecv()

  // Work with data from the Schema Registry
  messages := []proto.Message{
    &pb.Thing{}, // You can add more message types as needed
  }
  serde, err := schema.NewSerde(messages)
  if err != nil {
    log.Fatal().Err(err).Msg("Error initializing Serializer/Deserializer")
  }

  // Serialize the thing
  msg := pb.Thing{
    Foo: "bar"
  }
  bytes, err := serde.Encode(&msg)
  if err != nil {
    log.Fatal().Err(err).Msg("Error serializing")
  }
  err := kafkaClient.ProduceSync(ctx, &kgo.Record{
    Topic: "things",
    Value: bytes,
  }).FirstErr()
  if err != nil {
    log.Fatal().Err(err).Msg("Error producing")
  }

  // Deserialize the thing
  var record *kgo.Record
  msg = pb.Thing{}
  err = serde.Decode(record.Value, &msg)
  if err != nil {
    log.Fatal().Err(err).Msg("Error deserializing")
  }
}
```
