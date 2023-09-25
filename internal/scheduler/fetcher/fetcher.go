package fetcher

import (
	"context"
	"fmt"
	"time"

	"github.com/vincoll/vigie/internal/api/dbpgx"
	"github.com/vincoll/vigie/internal/scheduler/pulsar"
	"github.com/vincoll/vigie/pkg/business/core/probemgmt"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

type Fetcher struct {
	log           *zap.SugaredLogger
	db            *dbpgx.Client
	tracer        trace.Tracer
	OutgoingTests chan string
	done          chan bool
	//

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
				f.Fetch()
			}
		}
	}()
}

func (f *Fetcher) Fetch() {

	vts, err := f.probemgmt.GetByType(context.Background(), "icmp", time.Now())
	if err != nil {
		return 
	}
	fmt.Println(vts)

	f.log.Info("TICK !")
	f.OutgoingTests <- "jnhdl"
}

func (f *Fetcher) GracefulShutdown() {

	f.log.Infow(fmt.Sprintf("Shutdown fetcher service with frequency of %s", "x"), "component", "fetcher")
	f.done <- true

}
