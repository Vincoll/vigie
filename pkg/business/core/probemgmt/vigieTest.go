// Package probe provides an example of a core business API. Right now these
// calls are just wrapping the data/data layer. But at some point you will
// want to audit or something that isn't specific to the data/store layer.
package probemgmt

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/goccy/go-json"
	"github.com/google/uuid"
	"github.com/vincoll/vigie/internal/api/dbpgx"
	"github.com/vincoll/vigie/pkg/business/core/probemgmt/dbprobe"
	"github.com/vincoll/vigie/pkg/probe"
	"github.com/vincoll/vigie/pkg/probe/assertion"
	"github.com/vincoll/vigie/pkg/probe/icmp"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

// VigieTest is the expected struct received by the REST API
type VigieTest struct {
	Metadata   probe.Metadata         `json:"metadata"`
	Spec       interface{}            `json:"spec"`
	Assertions []*assertion.Assertion `json:"assertions"`
}

type VigieTestJSON struct {
	Metadata   probe.Metadata         `json:"metadata"`
	Spec       interface{}            `json:"spec"`
	Assertions []*assertion.Assertion `json:"assertions"`
}

// UnmarshalJSON vigieTest in a "Temp" Struct much closer than the REST Payload will be
// Conversion will be made to generate a clean VigieTest
func (vt *VigieTest) UnmarshalJSON(data []byte) error {

	var jsonTS VigieTestJSON
	if errjs := json.Unmarshal(data, &jsonTS); errjs != nil {
		return errjs
	}

	var err error
	*vt, err = jsonTS.toVigieTest()
	if err != nil {
		return fmt.Errorf("VigieTest is invalid: %s", err)
	}
	return nil
}

func (jvt VigieTestJSON) toVigieTest() (VigieTest, error) {

	// VigieTest init
	var vt VigieTest

	//
	// Metadata
	//
	if jvt.Metadata.Name == "" {
		return VigieTest{}, fmt.Errorf("name is missing or empty")
	}

	switch jvt.Metadata.Type {
	case
		"icmp",
		"tcp",
		"udp",
		"http",
		"This list will be registered elsewhere":

	default:
		return VigieTest{}, fmt.Errorf("type %q is invalid", jvt.Metadata.Type)
	}

	uuid, _ := uuid.NewRandom()
	vt.Metadata = jvt.Metadata
	vt.Metadata.UID = uint64(uuid.ID())

	//
	// Spec - Validation and Probe Init with default values
	//

	//var p2 probe.ProbeComplete
	var message proto.Message
	switch jvt.Metadata.Type {
	case "icmp":
		message = &icmp.Probe{}
	case "bar":
		message = &icmp.Probe{}
	}

	x := proto.Unmarshal([]byte(fmt.Sprint(jvt.Spec)), message)
	print(x)

	//
	// Assertions
	//

	vt.Assertions = jvt.Assertions

	return vt, nil
}

type Core struct {
	store dbprobe.ProbeDB
}

func NewCore(log *zap.SugaredLogger, db *dbpgx.Client) *Core {
	return &Core{
		store: dbprobe.NewProbeDB(log, db),
	}
}

// Set of error variables for CRUD operations.
var (
	ErrNotFoundProbe = errors.New("probe not found")
	ErrInvalidProbe  = errors.New("probe is not valid")
)

// Create inserts a new probe into the database.
func (c *Core) Create(ctx context.Context, nt *VigieTest, time time.Time) error {

	var prb dbprobe.ProbeTable

	err := c.store.XCreate3(ctx, prb)
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
		Metadata:   *pc.Metadata,
		Spec:       pc.Spec,
		Assertions: pc.Assertions,
	}

	return vt, nil

}
