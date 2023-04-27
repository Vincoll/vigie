package probemgmt

import (
	"time"

	"github.com/goccy/go-json"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/vincoll/vigie/pkg/probe"
	"google.golang.org/protobuf/proto"
)

// ProbeTable represent the structure we need for moving data
// between the app and the database.
type ProbeTable struct {
	ID         uuid.UUID        `db:"id"`
	ProbeType  string           `db:"probe_type"`
	Frequency  int              `db:"frequency"`
	Interval   pgtype.Interval  `db:"interval"`
	LastRun    pgtype.Timestamp `db:"last_run"`
	Probe_data []byte           `db:"probe_data"`
	Probe_json []byte           `db:"probe_json"`
}

// Converts VigieTestREST to a Struct ready to be insert in DB
func toProbeTable(vt VigieTest) (*ProbeTable, error) {

	pt := ProbeTable{
		ID:        uuid.UUID{},
		ProbeType: vt.Metadata.Type,
		Frequency: int(vt.Metadata.Frequency.Seconds),
		Interval: pgtype.Interval{
			Microseconds: vt.Metadata.Frequency.Seconds * 10000,
			Valid:        true,
		},
		LastRun:    pgtype.Timestamp{Time: time.Now().UTC(), Valid: true},
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
		Metadata:   vt.Metadata,
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
