package kafka

import (
	"fmt"
	"os"
	"strings"

	"github.com/twmb/franz-go/pkg/kgo"
)

func NewClient() (*kgo.Client, error) {
	brokers := os.Getenv("KAFKA_BROKERS")
	if brokers == "" {
		return nil, fmt.Errorf("KAFKA_BROKERS environment variable not set")
	}
	brokerList := strings.Split(brokers, ",")
	opts := []kgo.Opt{kgo.SeedBrokers(brokerList...)}

	return kgo.NewClient(opts...)
}
