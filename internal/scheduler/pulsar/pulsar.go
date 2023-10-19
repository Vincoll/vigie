package pulsar

import (
	"context"
	"fmt"
	"strings"
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
	IngoingTests chan []byte
	status       string
	logger       *zap.SugaredLogger
}

// New ...
func NewClient(ctx context.Context, conf ConfPulsar, logger *zap.SugaredLogger) (*PulsarClient, error) {

	logger.Infow(fmt.Sprintf("Initiate connection to %s as %v", conf.URL, conf), "component", "pulsar")

	_, span := otel.Tracer("vigie-boot").Start(ctx, "pulsar-init")
	defer span.End()

	client, err := pulsar.NewClient(pulsar.ClientOptions{
		URL:               conf.URL,
		OperationTimeout:  30 * time.Second,
		ConnectionTimeout: 30 * time.Second,
		MemoryLimitBytes:  64 * 1024 * 1024, // Unit: byte
		//logger:            logger.Desugar(),
	})
	if err != nil {
		span.SetStatus(codes.Error, fmt.Sprintf("Pulsar: Unable to connect to %s : %v", conf.URL, err))
		return nil, fmt.Errorf("Could not instantiate Pulsar client: %s ", err)
	}

	ch := make(chan []byte)

	pc := PulsarClient{client: client,
		IngoingTests: ch,
		status:       "running",
		logger:       logger,
		conf:         conf,
		producers:    make(map[string]pulsar.Producer),
	}

	err = pc.initProducers()
	if err != nil {
		return nil, err
	}

	go pc.Inject()

	span.SetStatus(codes.Ok, "Pulsar succesfully connected")
	logger.Infow(fmt.Sprintf("Pulsar connection established to %s as %v", conf.URL, conf), "component", "pulsar")

	return &pc, nil
}

func (p *PulsarClient) initProducers() error {

	vigieTopicPath := "vigie/test/"
	topics := []string{"test", "v0"}

	for _, topic := range topics {

		if !strings.Contains(topic, "/") {
			topic = vigieTopicPath + topic
		}
		p.addProducer(topic)
	}

	return nil
}

func (p *PulsarClient) addProducer(topic string) error {

	// Create a producer
	producer, err := p.client.CreateProducer(pulsar.ProducerOptions{
		Topic:       topic, //fmt.Sprintf("persistent://%s", topic),
		SendTimeout: 1 * time.Second,
	})
	if err != nil {
		return err
	}

	// Add each producer created to the map
	// They will be closed by the GracefulShutdown func
	p.producers[topic] = producer

	return nil

}

func (p *PulsarClient) Inject() {

	topicToSend := "vigie/test/test" // TEST DEBUG
	producerTopic := p.producers[topicToSend]
	ctx := context.Background()

	go func() {
		for {
			x := <-p.IngoingTests
			msgId, err := producerTopic.Send(ctx, &pulsar.ProducerMessage{
				Payload: x,
			})
			if err != nil {
				p.logger.Errorw(fmt.Sprintf("Failed to send Message (%s) to topic %s", msgId.String(), topicToSend), "component", "pulsar", "details", msgId)
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
