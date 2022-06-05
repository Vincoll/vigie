package core

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/jackc/pgtype"
	"github.com/vincoll/vigie/internal/api/dbsqlx"
	"github.com/vincoll/vigie/internal/api/health"
	"github.com/vincoll/vigie/internal/api/webapi"
	"github.com/vincoll/vigie/pkg/aaa/core/probe/db"
	"github.com/vincoll/vigie/pkg/load"
	"github.com/vincoll/vigie/pkg/teststruct"
	"github.com/vincoll/vigie/pkg/ticker"
	"go.uber.org/zap"
)

type vigieAPI struct {
	mu         sync.RWMutex
	TestSuites map[uint64]*teststruct.TestSuite

	ImportManager     load.ImportManager
	TickerPoolManager *ticker.TickerPoolManager
	incomingTests     chan map[uint64]*teststruct.TestSuite
	Health            health.AppHealthState
	logger            *zap.SugaredLogger
}

// StartVigieAPI Constructor of Vigie
func StartVigieAPI(apiConfig webapi.APIServerConfig, pgConfig dbsqlx.PGConfig, logger *zap.SugaredLogger) error {

	// Chans
	chanToScheduler := make(chan teststruct.Task)
	// Insert Chan Before (PoC) for now
	chanImportMgr := make(chan map[uint64]*teststruct.TestSuite)

	vAPI := vigieAPI{
		TestSuites:        map[uint64]*teststruct.TestSuite{},
		TickerPoolManager: ticker.NewTickerPoolManager(chanToScheduler),
		incomingTests:     chanImportMgr,
		Health:            health.AppHealthState{},
		logger:            logger,
	}

	// =========================================================================
	// Telemetry
	// Start Tracing Support

	// =========================================================================
	// Database
	//
	// Create connectivity to the database.
	dbc, err := dbsqlx.NewDBPool(pgConfig, logger)
	if err != nil {
		return fmt.Errorf("connecting to dbsqlx: %w", err)
	}

	// =========================================================================
	// HTTP API
	//
	// Start Vigie HTTP
	ws, err := webapi.NewHTTPServer(apiConfig, logger)

	// =========================================================================
	// AHS contains the global health state of the App
	//
	// It will
	// - Be linked to all components likely to break during Runtime (Eg: DB, Queue, Other Services,...)
	// - Start Technical HTTP Endpoints (Healthz, Metricz, ...)
	// - Report information on the runtime & exec
	// - Handle Graceful Shutdown
	health.NewAHS(apiConfig, ws, dbc, logger)

	// Placeholder
	vAPI.Health.HealthCheck()

	//////////////////////////////////

	y := db.NewStore(logger, dbc)

	interval := pgtype.Interval{}
	interval.Set(time.Second * 33)
	insert := db.ProbeDB{
		ID:         870,
		ProbeType:  "eza",
		Frequency:  9,
		Interval:   interval,
		LastRun:    pgtype.Timestamp{Time: time.Now().UTC(), Status: pgtype.Present},
		Probe_data: []byte(fmt.Sprint("cc")),
	}

	_, err = y.XQueryByID(context.Background(), "87696798")
	if err != nil {
		return err
	}

	// Wait Here
	select {}

	return nil
}
