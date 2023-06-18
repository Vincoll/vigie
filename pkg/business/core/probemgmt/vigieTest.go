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
	err = c.store.XCreate3(ctx, *dbvt)
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