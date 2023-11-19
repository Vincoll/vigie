package icmp

import "google.golang.org/protobuf/encoding/protojson"

type ProbeJSON struct {
	Host        string    `json:"Host,omitempty"`
	IPversion   Probe_IPv `json:"IPversion,omitempty"`
	PayloadSize int32     `json:"PayloadSize,omitempty"`
}

// UnmarshalJSON is a workaround to convert JSON data into a protobuf
// It wraps the protobuf "protojson" UnmarshalJSON method
func (x *Probe) UnmarshalJSON(data []byte) error {
	return protojson.Unmarshal(data, x)
}

func (x *Probe) MarshalJSON() ([]byte, error) {
	return protojson.Marshal(x)
}

func (x *Probe) ValidateAndInit() error {

	if x.GetPayloadSize() == 0 {
		x.PayloadSize = 32
	}

	return nil
}
