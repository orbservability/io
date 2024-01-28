package schema

import (
	"context"
	"fmt"
	"os"

	"github.com/twmb/franz-go/pkg/sr"
	"google.golang.org/protobuf/proto"
)

func NewSerde(messages []proto.Message) (*sr.Serde, error) {
	ctx := context.Background()

	// Ensure the URL is set for the Schema Registry
	schemaRegistryURL := os.Getenv("SCHEMA_REGISTRY_URL")
	if schemaRegistryURL == "" {
		return nil, fmt.Errorf("SCHEMA_REGISTRY_URL environment variable not set")
	}

	rcl, err := sr.NewClient(sr.URLs(schemaRegistryURL))
	if err != nil {
		return nil, err
	}

	var serde sr.Serde
	for _, message := range messages {
		schema, err := rcl.SchemaByVersion(ctx, string(proto.MessageName(message)), -1)
		if err != nil {
			return nil, err
		}

		// Register the message types with the serde
		serde.Register(
			schema.ID,
			message,
			sr.EncodeFn(func(v any) ([]byte, error) {
				return proto.Marshal(v.(proto.Message))
			}),
			sr.DecodeFn(func(b []byte, v any) error {
				return proto.Unmarshal(b, v.(proto.Message))
			}),
			sr.Index(0),
		)
	}

	return &serde, nil
}
