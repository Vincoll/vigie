package pulsar

import (
	"context"
	"fmt"
	"log"
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
	producers    map[string]pulsar.Producer
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

func (p *PulsarClient) initProducers() {

	topics := []string{"test", "test2"}

	for _, topic := range topics {
		p.addProducer(topic)
	}

}

func (p *PulsarClient) addProducer(topic string) {

	// Create a producer
	producer, err := p.client.CreateProducer(pulsar.ProducerOptions{
		Topic: topic,
	})
	if err != nil {
		log.Fatal(err)
	}

	// Add each producer created to the map
	// They will be closed by the GracefulShutdown func
	p.producers[topic] = producer

}

func (p *PulsarClient) Inject() {

	topicToSend := "test"
	producerTopic := p.producers[topicToSend]
	ctx := context.Background()

	go func() {
		for {

			x := <-p.IngoingTests
			msgId, err := producerTopic.Send(ctx, &pulsar.ProducerMessage{
				Payload: []byte(fmt.Sprintf(x)),
			})
			if err != nil {
				p.logger.Errorw(fmt.Sprintf("Faild to send Message (%s) to topic %s", msgId.String(), topicToSend), "component", "pulsar", "details", msgId)
			} else {
				p.logger.Infow(fmt.Sprintf("Message send to topic %s", topicToSend), "component", "pulsar")
			}
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
