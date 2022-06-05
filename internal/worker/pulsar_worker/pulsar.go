package pulsar_worker

import (
	"github.com/apache/pulsar-client-go/pulsar"
	"go.uber.org/zap"
	"time"
)

type ConfPulsar struct {
	Enable bool   `toml:"enable"`
	URL    string `toml:"url"`
}

type pulsarClient struct {
	conf ConfPulsar
}

// New ...
func NewClient(conf ConfPulsar) (*pulsar.Client, error) {

	client, err := pulsar.NewClient(pulsar.ClientOptions{
		URL:               conf.URL,
		OperationTimeout:  30 * time.Second,
		ConnectionTimeout: 30 * time.Second,
	})
	if err != nil {
		zap.S().Fatalf("Could not instantiate Pulsar client: %s ", err)
	}

	defer client.Close()
	return &client, nil
}
