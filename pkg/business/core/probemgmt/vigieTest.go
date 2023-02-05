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
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/vincoll/vigie/internal/api/dbpgx"
	"github.com/vincoll/vigie/pkg/business/core/probemgmt/dbprobe"
	"github.com/vincoll/vigie/pkg/probe"
	"github.com/vincoll/vigie/pkg/probe/assertion"
	"github.com/vincoll/vigie/pkg/probe/icmp"
	"go.uber.org/zap"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// VigieTest is the expected struct received by the REST API
// It's a less rigid struct that the full ProbeComplete Protobuf
type VigieTest struct {
	Metadata   probe.Metadata         `json:"metadata"`
	Spec       interface{}            `json:"spec"`
	Assertions []*assertion.Assertion `json:"assertions"`
}

// VigieTestJSON is a transition struct for UnMarshalJSON
// to Init and Validate data comming from the REST API
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

	var pc probe.ProbeComplete
	x := protojson.Unmarshal(data, &pc)
	print(x)

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

	x := protojson.Unmarshal([]byte(fmt.Sprint(jvt.Spec)), message)

	switch jvt.Metadata.Type {
	case "icmp":
		message.
	case "bar":
		var x icmp.Probe
	}

	var y probe.ProbeNotValidated

	y := x

	print(x)

	//
	// Assertions
	//

	vt.Assertions = jvt.Assertions

	return vt, nil
}

func (vt *VigieTest) ToProbeTable() (*dbprobe.ProbeTable, error) {

	pt := dbprobe.ProbeTable{
		ID:        uuid.UUID{},
		ProbeType: vt.Metadata.Type,
		Frequency: int(vt.Metadata.Frequency.Seconds),
		Interval: pgtype.Interval{
			Microseconds: vt.Metadata.Frequency.Seconds / 10000,
			Valid:        true,
		},
		LastRun:    pgtype.Timestamp{Time: time.Now().UTC()},
		Probe_data: nil,
		Probe_json: nil,
	}

	pc := probe.ProbeComplete{
		Metadata:   &vt.Metadata,
		Assertions: vt.Assertions,
		Spec:       nil,
	}

	//var p2 probe.ProbeComplete
	var message proto.Message
	switch vt.Metadata.Type {
	case "icmp":
		message = &icmp.Probe{}
	case "bar":
		message = &icmp.Probe{}
	default:
		return nil, fmt.Errorf("cannot ToProbeTable, %s type is unknown", vt.Metadata.Type)
	}
	err := proto.Unmarshal(pc.Spec.Value, message)
	if err != nil {
		return nil, err
	}

	// Set UUID before insert
	if vt.Metadata.UID == 0 {
		uuid, _ := uuid.NewRandom()
		vt.Metadata.UID = uint64(uuid.ID())
		pt.ID = uuid
	}
	pt.Probe_data, err = proto.Marshal(&pc)
	if err != nil {
		return nil, err
	}
	pt.Probe_json, err = json.Marshal(&pc)
	if err != nil {
		return nil, err
	}

	return &pt, nil

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
func (c *Core) Create(ctx context.Context, vt *VigieTest, time time.Time) error {

	// Need validation of VigieTest
	// TODO

	dbvt, err := vt.ToProbeTable()
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
		Metadata:   *pc.Metadata,
		Spec:       pc.Spec,
		Assertions: pc.Assertions,
	}

	return vt, nil

}
