package probemgmt

import (
	"fmt"
	"time"

	"github.com/goccy/go-json"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/vincoll/vigie/pkg/probe"
	"github.com/vincoll/vigie/pkg/probe/icmp"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
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

func toVigieTest(pt ProbeTable) (VigieTest, error) {
	return ByteToVigieTest(pt.Probe_data)
}

func ByteToVigieTest(probeData []byte) (VigieTest, error) {

	pc := probe.ProbeComplete{}

	if err := proto.Unmarshal(probeData, &pc); err != nil {
		return VigieTest{}, fmt.Errorf("could not deserialize anything: %s", err)
	}

	var prbType proto.Message
	switch pc.Metadata.Type {
	case "icmp":
		prbType = &icmp.Probe{}
	case "bar":
		prbType = &icmp.Probe{}
	}
	err := proto.Unmarshal(pc.Spec.Value, prbType)
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

func ByteToReadableVigieTest(pt ProbeTable) (VigieTest, error) {

	pc := probe.ProbeComplete{}

	if err := proto.Unmarshal(pt.Probe_data, &pc); err != nil {
		return VigieTest{}, fmt.Errorf("could not deserialize anything: %s", err)
	}

	var prbType proto.Message
	switch pc.Metadata.Type {
	case "icmp":
		prbType = &icmp.Probe{}
	case "bar":
		prbType = &icmp.Probe{}
	}
	err := proto.Unmarshal(pc.Spec.Value, prbType)
	if err != nil {
		return VigieTest{}, fmt.Errorf("could not protoUnmarshal: %s", err)

	}

	jsonBytes, err := protojson.Marshal(prbType)
	jsonString := string(jsonBytes)
	fmt.Println(jsonString)
	if err != nil {
		// handle error
	}

	vt := VigieTest{
		Metadata:   pc.Metadata,
		Spec:       &anypb.Any{TypeUrl: pc.Metadata.Type, Value: pc.Spec.Value},
		Assertions: pc.Assertions,
	}

	return vt, nil

}

/*
 https://ravina01997.medium.com/converting-interface-to-any-proto-and-vice-versa-in-golang-27badc3e23f1

https://dev.to/techschoolguru/go-generate-serialize-protobuf-message-4m7a
https://stackoverflow.com/questions/72381331/how-to-marshal-using-protojson-package-array-of-proto-to-json-in-golang
https://stackoverflow.com/questions/72473062/deserializing-external-json-payload-to-protobuf-any
*/
