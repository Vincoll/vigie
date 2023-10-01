package tracing

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"google.golang.org/grpc/credentials"
)

type OTelConfig struct {
	Enable      bool   `toml:"enable"`
	Url         string `toml:"url"`
	Port        string `toml:"port"`
	Env         string
	ServiceName string
	Version     string
}

type Client struct {
	tracer trace.Tracer
	meter  metric.Meter
	//numberOfExecutions metric.Int64Counter
	logger *zap.SugaredLogger
}

const (
	metricPrefix     = "custom.metric."
	numberOfExecName = metricPrefix + "number.of.exec"
	numberOfExecDesc = "Count the number of executions."
	heapMemoryName   = metricPrefix + "heap.memory"
	heapMemoryDesc   = "Reports heap memory utilization."
	httpsPreffix     = "https://"
)

func New(cfg OTelConfig, logger *zap.SugaredLogger) (*Client, error) {

	ctx := context.Background()

	// OpenTelemetry agent connectivity data
	_ = os.Getenv("EXPORTER_ENDPOINT")
	headers := os.Getenv("EXPORTER_HEADERS")
	headersMap := func(headers string) map[string]string {
		headersMap := make(map[string]string)
		if len(headers) > 0 {
			headerItems := strings.Split(headers, ",")
			for _, headerItem := range headerItems {
				parts := strings.Split(headerItem, "=")
				headersMap[parts[0]] = parts[1]
			}
		}
		return headersMap
	}(headers)

	c := Client{
		logger: logger,
	}

	fullEndpoint := fmt.Sprintf("%s:%s", cfg.Url, cfg.Port) //os.Getenv("EXPORTER_ENDPOINT")
	//headers := os.Getenv("EXPORTER_HEADERS")
	//	headersMap := map[string]string{"foo": "yop"}

	res0urce, err := resource.New(ctx,
		resource.WithFromEnv(),
		resource.WithProcess(),
		resource.WithTelemetrySDK(),
		resource.WithAttributes(
			semconv.ServiceNameKey.String(cfg.ServiceName),
			semconv.ServiceVersionKey.String(cfg.Version),
			semconv.TelemetrySDKLanguageGo,
		),
	)
	if err != nil {
		logger.Errorf("%s: %v", "failed to create resource", err)
	}

	// OTEL Collector / Exporter config -----------------------------------------

	// Initialize the tracer provider
	c.initTracer(ctx, fullEndpoint, headersMap, res0urce)

	// Initialize the meter provider
	//c.initMeter(ctx, fullEndpoint, headersMap, res0urce)

	return &c, nil

}

func (c *Client) GracefulShutdown() error {
	return nil
}

func (c *Client) initTracer(ctx context.Context, endpoint string,
	headersMap map[string]string, res0urce *resource.Resource) {

	traceOpts := []otlptracegrpc.Option{
		otlptracegrpc.WithTimeout(5 * time.Second),
	}
	if strings.Contains(endpoint, httpsPreffix) {
		endpoint = strings.ReplaceAll(endpoint, httpsPreffix, "")
		traceOpts = append(traceOpts, otlptracegrpc.WithHeaders(headersMap))
		traceOpts = append(traceOpts, otlptracegrpc.WithTLSCredentials(credentials.NewTLS(&tls.Config{})))
	} else {
		endpoint = strings.ReplaceAll(endpoint, "http://", "")
		traceOpts = append(traceOpts, otlptracegrpc.WithInsecure())
	}
	traceOpts = append(traceOpts, otlptracegrpc.WithEndpoint(endpoint))

	traceExporter, err := otlptracegrpc.New(ctx, traceOpts...)
	if err != nil {
		log.Fatalf("%s: %v", "failed to create exporter", err)
	}

	otel.SetTracerProvider(sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithResource(res0urce),
		sdktrace.WithSpanProcessor(
			sdktrace.NewBatchSpanProcessor(traceExporter)),
	))

	otel.SetTextMapPropagator(
		propagation.NewCompositeTextMapPropagator(
			propagation.Baggage{},
			propagation.TraceContext{},
		),
	)

	// c.tracer = otel.Tracer("io.opentelemetry.traces.hello")

}

/*
func (c *Client) initMeter(ctx context.Context, endpoint string,
	headersMap map[string]string, res0urce *resource.Resource) {

	metricOpts := []otlpmetricgrpc.Option{
		otlpmetricgrpc.WithTimeout(5 * time.Second),
	}
	if strings.Contains(endpoint, httpsPreffix) {
		endpoint = strings.ReplaceAll(endpoint, httpsPreffix, "")
		metricOpts = append(metricOpts, otlpmetricgrpc.WithHeaders(headersMap))
		metricOpts = append(metricOpts, otlpmetricgrpc.WithTLSCredentials(credentials.NewTLS(&tls.Config{})))
	} else {
		endpoint = strings.ReplaceAll(endpoint, "http://", "")
		metricOpts = append(metricOpts, otlpmetricgrpc.WithInsecure())
	}
	metricOpts = append(metricOpts, otlpmetricgrpc.WithEndpoint(endpoint))

	metricExporter, err := otlpmetricgrpc.New(ctx, metricOpts...)
	if err != nil {
		log.Fatalf("%s: %v", "failed to create exporter", err)
	}

	pusher := controller.New(
		processor.NewFactory(
			simple.NewWithHistogramDistribution(),
			metricExporter,
		),
		controller.WithResource(res0urce),
		controller.WithExporter(metricExporter),
		controller.WithCollectPeriod(5*time.Second),
	)

	err = pusher.Start(ctx)
	if err != nil {
		log.Fatalf("%s: %v", "failed to start the pusher", err)
	}

	global.SetMeterProvider(pusher)
	c.meter = global.Meter("io.opentelemetry.metrics.hello")

}
*/
