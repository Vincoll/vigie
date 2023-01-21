package core

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/golang/protobuf/ptypes"
	"github.com/jackc/pgtype"
	"github.com/vincoll/vigie/internal/api/conf"
	"github.com/vincoll/vigie/internal/api/dbpgx"
	"github.com/vincoll/vigie/internal/api/health"
	"github.com/vincoll/vigie/internal/api/webapi"
	"github.com/vincoll/vigie/pkg/business/core/probemgmt/dbprobe"
	"github.com/vincoll/vigie/pkg/load"
	"github.com/vincoll/vigie/pkg/probe"
	"github.com/vincoll/vigie/pkg/probe/icmp"
	"github.com/vincoll/vigie/pkg/teststruct"
	"github.com/vincoll/vigie/pkg/ticker"
	"github.com/vincoll/vigie/pkg/tracing"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
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

// NewVigieAPI Constructor of Vigie
func NewVigieAPI(appCfg conf.VigieAPIConf, logger *zap.SugaredLogger) error {

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
	//
	// Start Tracing Support
	otClient, err := tracing.New(appCfg.OTel, logger)
	if err != nil {
		return fmt.Errorf("fail to contact the OpenTelemetry endpoint: %w", err)
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
	// HTTP API
	//
	// Start Vigie HTTP
	ws, err := webapi.NewHTTPServer(ctxSpan, appCfg.HTTP, appCfg.Environment, logger, dbc)
	if err != nil {
		return fmt.Errorf("fail to load HTTP Server: %w", err)
	}

	// =========================================================================
	// AHS contains the global health state of the App
	//
	// It will
	// - Be linked to all components likely to break during Runtime (Eg: DB, Queue, Other Services,...)
	// - Start Technical HTTP Endpoints (Healthz, Metricz, ...)
	// - Report information on the runtime & exec
	// - Handle Graceful Shutdown
	health.NewAHS(appCfg.HTTP, ws, dbc, otClient, logger)

	// Placeholder
	vAPI.Health.HealthCheck()

	generateThings(logger, dbc)

	bootSpan.SetStatus(codes.Ok, "App Started Successfully")
	bootSpan.End()

	// Wait Here
	select {}
	return nil
}

func generateThings(logger *zap.SugaredLogger, dbc *dbpgx.Client) error {

	//////////////////////////////////

	met := probe.Metadata{
		UID:         0,
		Name:        "Meta-Name-ICMP",
		Type:        "icmp",
		LastAttempt: timestamppb.New(time.Now()),
		Frequency:   durationpb.New(time.Second * 33),
	}

	i := icmp.New()
	i.Host = "127.0.0.1"
	serialized, err := proto.Marshal(i)
	if err != nil {
	}

	probex := probe.ProbeComplete{
		Metadata:   &met,
		Assertions: nil,
		Spec: &anypb.Any{
			TypeUrl: "icmp",
			Value:   serialized,
		},
	}

	d, err := proto.Marshal(&probex)

	// unmarshal to simulate coming off the wire
	var p2 probe.ProbeComplete

	if err := proto.Unmarshal(d, &p2); err != nil {
		fmt.Sprintf("could not deserialize anything")
	}

	var message proto.Message
	switch p2.Spec.TypeUrl {
	case "icmp":
		message = &icmp.Probe{}
	case "bar":
		message = &icmp.Probe{}
	}
	x := proto.Unmarshal(p2.Spec.Value, message)

	fmt.Sprintf("%s", x)

	// unmarshal the timestamp
	var i2 icmp.Probe

	if err := ptypes.UnmarshalAny(p2.Spec, &i2); err != nil {
		fmt.Sprintf("Could not unmarshal timestamp from anything field: %s", err)
	}

	dj, err := json.Marshal(&probex)

	y := dbprobe.NewProbeDB(logger, dbc)

	interval := pgtype.Interval{}
	interval.Set(time.Second * 33)

	for i := 1; i < 3; i++ {

		insert := dbprobe.ProbeTable{
			ProbeType: "icmp",
			Interval:  interval,
			LastRun:   pgtype.Timestamp{Time: time.Now().UTC(), Status: pgtype.Present},

			Probe_data: d,
			Probe_json: dj,
		}

		err = y.XCreate3(context.Background(), insert)
		if err != nil {
			return err
		}
	}

	/*
		_, err = y.XQueryByID(context.Background(), "87696798")
		if err != nil {
			return err
		}
	*/
	return nil
}
