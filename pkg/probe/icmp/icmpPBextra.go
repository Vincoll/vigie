package icmp

import "google.golang.org/protobuf/runtime/protoimpl"

type ProbeJSON struct {
	Host        string    `json:"Host,omitempty"`
	IPversion   Probe_IPv `json:"IPversion,omitempty"`
	PayloadSize int32     `json:"PayloadSize,omitempty"`
}

func (x *Probe) ImportJSON() {
	*x = Probe{}
	if protoimpl.UnsafeEnabled {
		mi := &file_icmp_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}
