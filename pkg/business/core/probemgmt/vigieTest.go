// Package probe provides an example of a core business API. Right now these
// calls are just wrapping the data/data layer. But at some point you will
// want to audit or something that isn't specific to the data/store layer.
package probemgmt

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

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
	"google.golang.org/protobuf/types/known/anypb"
)

// VigieTestREST is the expected struct received by the REST API
// It's a less rigid struct that the full ProbeComplete Protobuf
type VigieTestREST struct {
	Metadata   probe.Metadata         `json:"metadata"`
	Spec       *anypb.Any             `json:"spec"`
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
// Conversion will be made to generate a clean VigieTestREST
func (vt *VigieTestREST) UnmarshalJSON(data []byte) error {

	var jsonTS VigieTestJSON
	if errjs := json.Unmarshal(data, &jsonTS); errjs != nil {
		return errjs
	}
	/*
		var pc probe.ProbeComplete
		x := protojson.Unmarshal(data, &pc)
		print(x)
	*/
	var err error
	*vt, err = jsonTS.toVigieTest()
	if err != nil {
		return fmt.Errorf("VigieTestREST is invalid: %s", err)
	}
	return nil
}

func (jvt VigieTestJSON) toVigieTest() (VigieTestREST, error) {

	// VigieTestREST init
	var vt VigieTestREST

	//
	// Metadata
	//
	if jvt.Metadata.Name == "" {
		return VigieTestREST{}, fmt.Errorf("name is missing or empty")
	}
	vt.Metadata = jvt.Metadata

	//
	// Spec - Validation and Probe Init with default values
	//

	var message proto.Message
	// 		"This list will be registered elsewhere":
	// This must be re-factored
	switch jvt.Metadata.Type {
	case "icmp":
		m := icmp.New()

		x, err := json.Marshal(jvt.Spec)
		if err != nil {
			return VigieTestREST{}, err
		}
		// Populate the ICMP Spec with the data from the REST Payload
		err = protojson.Unmarshal(x, m)
		if err != nil {
			return VigieTestREST{}, err
		}
		// Validate and Init the Probe
		err = m.ValidateAndInit()
		if err != nil {
			return VigieTestREST{}, err
		}
		message = m
	case "tcp":
		message = &icmp.Probe{}
	default:
		return VigieTestREST{}, fmt.Errorf("type %q is invalid", jvt.Metadata.Type)
	}

	// convert vt.Spec to anypb.any
	spec, err := anypb.New(message)
	if err != nil {
		return VigieTestREST{}, err
	}

	vt.Spec = spec

	//
	// Assertions
	//

	vt.Assertions = jvt.Assertions

	return vt, nil
}

// Converts VigieTestREST to a a Struct ready to be insert in DB
func (vt *VigieTestREST) ToProbeTable() (*dbprobe.ProbeTable, error) {

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

	// Set UUID before insert
	if vt.Metadata.UID == 0 {
		genUuid, _ := uuid.NewRandom()
		vt.Metadata.UID = uint64(genUuid.ID())
		pt.ID = genUuid
	}

	// pc will be used in Probe_data and Probe_json
	// probe_data as pure Protobuf
	// probe_json as byte but JSON encoded
	pc := probe.ProbeComplete{
		Metadata:   &vt.Metadata,
		Assertions: vt.Assertions,
		Spec:       vt.Spec,
	}

	var err error
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
func (c *Core) Create(ctx context.Context, vt *VigieTestREST) error {

	// Need validation of VigieTestREST
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
func (c *Core) GetByID(ctx context.Context, id string, time time.Time) (VigieTestREST, error) {

	// Get the entire row
	pt, err := c.store.QueryByID(ctx, id)
	if err != nil {
		return VigieTestREST{}, err
	}

	pc := probe.ProbeComplete{}

	if err := proto.Unmarshal(pt.Probe_data, &pc); err != nil {
		return VigieTestREST{}, fmt.Errorf("could not deserialize anything: %s", err)
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
		return VigieTestREST{}, fmt.Errorf("could not protoUnmarshal: %s", err)

	}

	vt := VigieTestREST{
		Metadata:   *pc.Metadata,
		Spec:       pc.Spec,
		Assertions: pc.Assertions,
	}

	return vt, nil

}
