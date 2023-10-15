package pulsar

import (
	"context"
	"fmt"
	"time"

	"github.com/apache/pulsar-client-go/pulsar"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"

	"go.uber.org/zap"
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
	logger       *zap.SugaredLogger
}

// New ...
func NewClient(ctx context.Context, conf ConfPulsar, logger *zap.SugaredLogger) (*PulsarClient, error) {

	_, span := otel.Tracer("vigie-boot").Start(ctx, "pulsar-init")
	defer span.End()

	client, err := pulsar.NewClient(pulsar.ClientOptions{
		URL:               conf.URL,
		OperationTimeout:  30 * time.Second,
		ConnectionTimeout: 30 * time.Second,
		MemoryLimitBytes:  64 * 1024 * 1024, // Unit: byte

		ListenerName: "scheduler",
	})
	if err != nil {
		span.SetStatus(codes.Error, fmt.Sprintf("Pulsar: Unable to connect to %s : %v", conf.URL, err))
		return nil, fmt.Errorf("Could not instantiate Pulsar client: %s ", err)
	}

	ch := make(chan string)

	pc := PulsarClient{client: client, IngoingTests: ch, status: "running"}

	go pc.Inject()

	span.SetStatus(codes.Ok, "Pulsar succesfully connected")
	logger.Infow(fmt.Sprintf("Pulsar connection established to %s as %v", conf.URL, conf), "component", "pulsar")

	return &pc, nil
}

func (p *PulsarClient) Inject() {

	go func() {
		for {
			fmt.Println(<-p.IngoingTests)
		}
	}()

}

func (p *PulsarClient) GracefulShutdown() error {

	p.logger.Infow("Pulsar Client is gracefully shutting down")

	// Set app UnHealthy
	p.status = "ShuttingDown"

	p.client.Close()
	return nil
}
