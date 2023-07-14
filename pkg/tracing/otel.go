package tracing

import (
	"context"
	"crypto/tls"
	"fmt"
	"strings"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
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
	serviceName      = "hello-app"
	serviceVersion   = "v1.0.0"
	metricPrefix     = "custom.metric."
	numberOfExecName = metricPrefix + "number.of.exec"
	numberOfExecDesc = "Count the number of executions."
	heapMemoryName   = metricPrefix + "heap.memory"
	heapMemoryDesc   = "Reports heap memory utilization."
	httpsPreffix     = "https://"

	service     = "trace-demo"
	environment = "production"
	id          = 1
)

func New(cfg OTelConfig, logger *zap.SugaredLogger) (*Client, error) {

	ctx := context.Background()

	c := Client{
		logger: logger,
	}

	endpoint := cfg.Url //os.Getenv("EXPORTER_ENDPOINT")
	endpoint = "localhost:4317"
	//headers := os.Getenv("EXPORTER_HEADERS")
	//	headersMap := map[string]string{"foo": "yop"}

	res, err := resource.New(ctx,
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

	traceOpts := []otlptracegrpc.Option{
		otlptracegrpc.WithTimeout(5 * time.Second),
	}
	if strings.Contains(endpoint, httpsPreffix) {
		endpoint = strings.ReplaceAll(endpoint, httpsPreffix, "")
		//	traceOpts = append(traceOpts, otlptracegrpc.WithHeaders(headersMap))
		traceOpts = append(traceOpts, otlptracegrpc.WithTLSCredentials(credentials.NewTLS(&tls.Config{})))
	} else {
		traceOpts = append(traceOpts, otlptracegrpc.WithInsecure())
	}
	traceOpts = append(traceOpts, otlptracegrpc.WithEndpoint(endpoint))

	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, endpoint, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	if err != nil {
		return nil, fmt.Errorf("failed to create gRPC connection to collector: %w", err)
	}

	_, err = otlptracegrpc.New(ctx, otlptracegrpc.WithGRPCConn(conn))
	if err != nil {
		return nil, fmt.Errorf("failed to create trace exporter: %w", err)
	}

	// ------------------------------------------------------

	// Set up a trace exporter

	traceClient := otlptracegrpc.NewClient(
		otlptracegrpc.WithInsecure(),
		otlptracegrpc.WithEndpoint("localhost:4317"),
		otlptracegrpc.WithDialOption(grpc.WithBlock()))
	traceExporter, err := otlptrace.New(ctx, traceClient)

	//traceExporter, err := otlptracegrpc.New(ctx, otlptracegrpc.WithGRPCConn(conn))
	if err != nil {
		return nil, fmt.Errorf("failed to create trace exporter: %w", err)
	}

	// _______________________________________________________________

	// Register the trace exporter with a TracerProvider, using a batch
	// span processor to aggregate spans before export.

	bsp := sdktrace.NewBatchSpanProcessor(traceExporter)
	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithResource(res),
		sdktrace.WithSpanProcessor(bsp),
	)

	// Set GLOBAL propagator to tracecontext (the default is no-op).
	otel.SetTracerProvider(tracerProvider)
	otel.SetTextMapPropagator(propagation.TraceContext{})

	//  Silencing logs
	otel.SetErrorHandler(otel.ErrorHandlerFunc(func(err error) {
		// ignore the error
	}))

	/*
		tracer := otel.Tracer("test-tracer")

		// work begins
		// Attributes represent additional key-value descriptors that can be bound
		// to a metric observer or recorder.
		ctx, span := otel.Tracer("test-tracer").Start(
			ctx,
			"CollectorExporter-Example",
			trace.WithAttributes([]attribute.KeyValue{}...))F
		defer span.End()
		for i := 0; i < 2; i++ {
			_, iSpan := tracer.Start(ctx, fmt.Sprintf("OUT-Sample-%d", i))
			fmt.Printf("Doing really hard work (%d / 10)\n", i+1)

			<-time.After(time.Second)
			iSpan.End()
		}
	*/

	// Initialize the meter provider
	//c.initMeter(ctx, endpoint, headersMap, res)

	// Create the metrics
	//createMetrics()

	return &c, nil

}

func (c *Client) GracefulShutdown() error {

	return nil
}

/*
func (c *Client) initMeter(ctx context.Context, endpoint string, headersMap map[string]string, res0urce *resource.Resource) {

	metricOpts := []otlpmetricgrpc.Option{
		otlpmetricgrpc.WithTimeout(5 * time.Second),
	}
	if strings.Contains(endpoint, httpsPreffix) {
		endpoint = strings.ReplaceAll(endpoint, httpsPreffix, "")
		metricOpts = append(metricOpts, otlpmetricgrpc.WithHeaders(headersMap))
		metricOpts = append(metricOpts, otlpmetricgrpc.WithTLSCredentials(credentials.NewTLS(&tls.Config{})))
	} else {
		metricOpts = append(metricOpts, otlpmetricgrpc.WithInsecure())
	}
	metricOpts = append(metricOpts, otlpmetricgrpc.WithEndpoint(endpoint))

	metricExporter, err := otlpmetricgrpc.New(ctx, metricOpts...)
	if err != nil {
		c.logger.Errorf("%s: %v", "failed to create exporter", err)
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
		c.logger.Errorf("%s: %v", "failed to start the pusher", err)
	}

	global.SetMeterProvider(pusher)
	c.meter = global.Meter("io.opentelemetry.metrics.hello")

}

func (c *Client) GracefulShutdown() error {

	return nil
}
*/
/*
func createMetrics() {

	// Metric to be updated manually
	numberOfExecutions = metric.Must(meter).
		NewInt64Counter(
			numberOfExecName,
			metric.WithDescription(numberOfExecDesc),
		)

	// Metric to be updated automatically
	_ = metric.Must(meter).
		NewInt64CounterObserver(
			heapMemoryName,
			func(_ context.Context, result metric.Int64ObserverResult) {
				var mem runtime.MemStats
				runtime.ReadMemStats(&mem)
				result.Observe(int64(mem.HeapAlloc),
					attribute.String(heapMemoryName,
						heapMemoryDesc))
			},
			metric.WithDescription(heapMemoryDesc))

}
*/
