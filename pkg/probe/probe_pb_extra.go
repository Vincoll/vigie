package probe

import "google.golang.org/protobuf/encoding/protojson"

type ProbeNotValidated interface {
	ValidateAndInit() error
}

type ProbeNotVal any

// UnmarshalJSON converts JSON data into a Providers.Polygon.ArrayResponse
// https://stackoverflow.com/questions/72473062/deserializing-external-json-payload-to-protobuf-any
func (x *ProbeComplete) UnmarshalJSON(data []byte) error {
	return protojson.Unmarshal(data, x)
}

func (x *ProbeComplete) MarshalJSON() ([]byte, error) {
	return protojson.Marshal(x)
}
