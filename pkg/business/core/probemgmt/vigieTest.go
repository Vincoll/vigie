// Package probe provides an example of a core business API. Right now these
// calls are just wrapping the data/data layer. But at some point you will
// want to audit or something that isn't specific to the data/store layer.
package probemgmt

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/vincoll/vigie/internal/api/dbpgx"
	"github.com/vincoll/vigie/pkg/probe"
	"github.com/vincoll/vigie/pkg/probe/assertion"
	"github.com/vincoll/vigie/pkg/probe/icmp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

type VigieTest struct {
	Metadata   *probe.Metadata        `json:"metadata"`
	Spec       *anypb.Any             `json:"spec"`
	Assertions []*assertion.Assertion `json:"assertions"`
}

type Core struct {
	store ProbeDB
}

func NewCore(log *zap.SugaredLogger, db *dbpgx.Client) *Core {
	return &Core{
		store: NewProbeDB(log, db),
	}
}

// Set of error variables for CRUD operations.
var (
	ErrNotFoundProbe = errors.New("probe not found")
	ErrInvalidProbe  = errors.New("probe is not valid")
	ErrDBUnavailable = errors.New("database unavailable")
)

// Create inserts a new probe into the database.
func (c *Core) Create(ctx context.Context, vt *VigieTest) error {

	// Need validation of VigieTestREST
	// TODO

	dbvt, err := toProbeTable(*vt)
	if err != nil {
		return err
	}
	err = c.store.Create(ctx, *dbvt)
	if err != nil {
		return err
	}

	return nil
}

// GetByID Get a test by his ID from the database.
func (c *Core) GetByID(ctx context.Context, id string, time time.Time) (VigieTest, error) {

	// Get the entire row
	pt, err := c.store.QueryByID(ctx, id)
	if err != nil {
		return VigieTest{}, err
	}

	pc := probe.ProbeComplete{}

	if err := proto.Unmarshal(pt.Probe_data, &pc); err != nil {
		return VigieTest{}, fmt.Errorf("could not deserialize anything: %s", err)
	}

	var prbType proto.Message
	switch pt.ProbeType {
	case "icmp":
		prbType = &icmp.Probe{}
	case "bar":
		prbType = &icmp.Probe{}
	}
	err = proto.Unmarshal(pc.Spec.Value, prbType)
	if err != nil {
		return VigieTest{}, fmt.Errorf("could not protoUnmarshal: %s", err)

	}

	vt := VigieTest{
		Metadata:   pc.Metadata,
		Spec:       pc.Spec,
		Assertions: pc.Assertions,
	}

	return vt, nil

}

// GetByType Returns tests for a given type from the database.
func (c *Core) GetByType(ctx context.Context, probeType string, time time.Time) ([]VigieTest, error) {

	// Get the entire row
	pt, err := c.store.QueryByType(ctx, probeType)
	if err != nil {
		return nil, err
	}

	vts := make([]VigieTest, 0, len(pt))
	//

	for _, p := range pt {

		var pcs probe.ProbeComplete

		if err := proto.Unmarshal(p.Probe_data, &pcs); err != nil {
			return nil, fmt.Errorf("could not deserialize anything: %s", err)
		}

		var prbType proto.Message
		switch p.ProbeType {
		case "icmp":
			prbType = &icmp.Probe{}
		case "bar":
			prbType = &icmp.Probe{}
		}
		err = proto.Unmarshal(pcs.Spec.Value, prbType)
		if err != nil {
			return nil, fmt.Errorf("could not protoUnmarshal: %s", err)

		}

		vt := VigieTest{
			Metadata:   pcs.Metadata,
			Spec:       pcs.Spec,
			Assertions: pcs.Assertions,
		}
		vts = append(vts, vt)
	}
	return vts, nil

}

// GetTestsPastInterval  Returns tests requiring to be executed in the past interval.
func (c *Core) GetTestsPastInterval(ctx context.Context, probeType string, interval time.Duration) ([]VigieTest, error) {

	_, spanDB := otel.Tracer("fetch-get1mTests").Start(ctx, "query-db")
	// Get the entire row
	pt, err := c.store.QueryPastInterval(ctx, probeType, interval)
	if err != nil {
		return nil, err
	}
	spanDB.SetStatus(codes.Ok, "DB query OK")
	spanDB.End()

	_, spanProcess := otel.Tracer("fetch-get1mTests").Start(ctx, "process-data-from-db")
	defer spanProcess.End()

	vts := make([]VigieTest, 0, len(pt))

	for _, p := range pt {

		var pcs probe.ProbeComplete

		if err := proto.Unmarshal(p.Probe_data, &pcs); err != nil {
			return nil, fmt.Errorf("could not deserialize anything: %s", err)
		}

		var prbType proto.Message
		switch p.ProbeType {
		case "icmp":
			prbType = &icmp.Probe{}
		case "bar":
			prbType = &icmp.Probe{}
		}
		err = proto.Unmarshal(pcs.Spec.Value, prbType)
		if err != nil {
			return nil, fmt.Errorf("could not protoUnmarshal: %s", err)

		}

		vt := VigieTest{
			Metadata:   pcs.Metadata,
			Spec:       pcs.Spec,
			Assertions: pcs.Assertions,
		}
		vts = append(vts, vt)
	}
	spanProcess.SetStatus(codes.Ok, "Process data OK")
	return vts, nil

}

// GetTestsPastInterval  Returns tests requiring to be executed in the past interval.
func (c *Core) GetTestsPastIntervalProbeData(ctx context.Context, probeType string, interval time.Duration) ([][]byte, error) {

	_, spanDB := otel.Tracer("fetch-get1mTests-raw").Start(ctx, "query-db")
	// Get the entire row
	pt, err := c.store.QueryPastInterval(ctx, probeType, interval)
	if err != nil {
		return nil, err
	}
	spanDB.SetStatus(codes.Ok, "DB query OK")
	spanDB.End()

	_, spanProcess := otel.Tracer("fetch-get1mTests-raw").Start(ctx, "process-data-from-db-raw")
	defer spanProcess.End()

	vts := make([][]byte, 0, len(pt))

	for _, p := range pt {
		vts = append(vts, p.Probe_data)
	}
	spanProcess.SetStatus(codes.Ok, "Process data OK")
	return vts, nil

}
