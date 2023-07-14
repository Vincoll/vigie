package pulsar

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/apache/pulsar-client-go/pulsar"
)

type ConfPulsar struct {
	Enable bool   `toml:"enable"`
	URL    string `toml:"url"`
}

type PulsarClient struct {
	conf         ConfPulsar
	client       pulsar.Client
	IngoingTests chan string
	status       string
}

// New ...
func NewClient(ctx context.Context, conf ConfPulsar, logger *zap.SugaredLogger) (*PulsarClient, error) {

	client, err := pulsar.NewClient(pulsar.ClientOptions{
		URL:               conf.URL,
		OperationTimeout:  30 * time.Second,
		ConnectionTimeout: 30 * time.Second,
	})
	if err != nil {
		logger.Fatalf("Could not instantiate Pulsar client: %s ", err)
	}

	ch := make(chan string)

	pc := PulsarClient{client: client, IngoingTests: ch, status: "running"}

	go pc.Inject()

	return &pc, nil
}

func (p *PulsarClient) Inject() {

	// Set app UnHealthy
	go func() {
		for {
			fmt.Println(<-p.IngoingTests)
		}
	}()

}

func (p *PulsarClient) GracefulShutdown() error {

	// Set app UnHealthy
	p.status = "ShuttingDown"

	p.client.Close()
	return nil
}
