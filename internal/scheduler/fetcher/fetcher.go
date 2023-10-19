package fetcher

import (
	"context"
	"fmt"
	"time"

	"github.com/vincoll/vigie/internal/api/dbpgx"
	"github.com/vincoll/vigie/internal/scheduler/pulsar"
	"github.com/vincoll/vigie/pkg/business/core/probemgmt"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

type Fetcher struct {
	log           *zap.SugaredLogger
	db            *dbpgx.Client
	tracer        trace.Tracer
	OutgoingTests chan []byte
	done          chan bool

	probemgmt *probemgmt.Core
}

func NewFetcher(ctx context.Context, pulc *pulsar.PulsarClient, log *zap.SugaredLogger, db *dbpgx.Client, tracer trace.Tracer) (*Fetcher, error) {

	pm := probemgmt.NewCore(log, db)

	f := Fetcher{
		log:           log,
		db:            db,
		OutgoingTests: pulc.IngoingTests,
		tracer:        tracer,
		probemgmt:     pm,
	}

	f.Start()

	return &f, nil

}

func (f *Fetcher) Start() {

	freq := time.Duration(1) * time.Second
	f.log.Infow(fmt.Sprintf("Starting fetcher service with frequency of %s", freq), "component", "fetcher")

	done := make(chan bool)

	go func() {

		importTicker := time.NewTicker(freq)

		for {
			select {
			case <-done:
				return

			case <-importTicker.C:
				f.get1mTests()
			}
		}
	}()
}

func (f *Fetcher) GracefulShutdown() {

	f.log.Infow(fmt.Sprintf("Shutdown fetcher service with frequency of %s", "x"), "component", "fetcher")
	f.done <- true

}

func (f *Fetcher) get1mTests() error {

	tracer := otel.Tracer("fetch-get1mTests")
	ctxSpan, get1mSpan := tracer.Start(context.Background(), "fetch-get1mTests")

	_, _ = f.probemgmt.GetTestsPastInterval(ctxSpan, "icmp", time.Minute)
	vts, err := f.probemgmt.GetTestsPastIntervalProbeData(ctxSpan, "icmp", time.Minute)
	if err != nil {
		return err
	}
	_ = vts

	for _, v := range vts {

		f.OutgoingTests <- v

	}
	get1mSpan.SetStatus(codes.Ok, "get1mTests Successfully")
	get1mSpan.End()

	return nil
}
