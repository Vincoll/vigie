package core

import (
	"context"
	"fmt"
	"sync"

	"github.com/vincoll/vigie/internal/api/dbpgx"
	"github.com/vincoll/vigie/internal/scheduler/conf"
	"github.com/vincoll/vigie/internal/scheduler/fetcher"
	"github.com/vincoll/vigie/internal/scheduler/health"
	"github.com/vincoll/vigie/internal/scheduler/pulsar"
	"github.com/vincoll/vigie/pkg/tracing"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.uber.org/zap"
)

type vigieScheduler struct {
	mu     sync.RWMutex
	Health health.AppHealthState
	logger *zap.SugaredLogger
}

// NewVigieScheduler Constructor of Vigie
func NewVigieScheduler(appCfg conf.VigieSchedulerConf, logger *zap.SugaredLogger) error {

	vAPI := vigieScheduler{
		Health: health.AppHealthState{},
		logger: logger,
	}

	// =========================================================================
	// Telemetry
	//
	// Start Tracing Support
	otClient, err := tracing.New(appCfg.OTel, logger)
	if err != nil {
		logger.Errorf("fail to contact the OpenTelemetry endpoint: %s", err)
		// return fmt.Errorf("fail to contact the OpenTelemetry endpoint: %w", err)
	}

	// Start to Trace the boot of vigie-agi
	tracer := otel.Tracer("vigie-boot")
	ctxSpan, bootSpan := tracer.Start(context.Background(), "vigie-boot")

	// =========================================================================
	// Database
	//
	// Create connectivity to the database.
	dbc, err := dbpgx.NewDBPool(ctxSpan, appCfg.PG, logger)
	if err != nil {
		return fmt.Errorf("fail to connecting to dbsqlx: %w", err)
	}

	// =========================================================================
	// Pulsar
	//
	// Create connectivity to the Pulsar
	pulc, err := pulsar.NewClient(ctxSpan, appCfg.Pulsar, logger)
	if err != nil {
		return fmt.Errorf("fail to connecting to dbsqlx: %w", err)
	}

	// =========================================================================
	// Fetcher
	//
	// Fetch tests from DB
	ftchr, err := fetcher.NewFetcher(ctxSpan, pulc, logger, dbc, tracer)
	if err != nil {
		return fmt.Errorf("fail to start the fetcher: %w", err)
	}

	// =========================================================================
	// AHS contains the global health state of the App
	//
	// It will
	// - Be linked to all components likely to break during Runtime (Eg: DB, Queue, Other Services,...)
	// - Start Technical HTTP Endpoints (Healthz, Metricz, ...)
	// - Report information on the runtime & exec
	// - Handle Graceful Shutdown
	health.NewAHS(appCfg.HTTP, dbc, pulc, ftchr, otClient, logger)

	// Placeholder
	vAPI.Health.HealthCheck()

	//generateThings(logger, dbc)

	bootSpan.SetStatus(codes.Ok, "App Started Successfully")
	bootSpan.End()

	// Wait Here
	select {}
	return nil
}
