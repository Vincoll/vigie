// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.0
// 	protoc        v4.25.0
// source: tcp.proto

package tcp

import (
	probe "github.com/vincoll/vigie/pkg/probe"
	assertion "github.com/vincoll/vigie/pkg/probe/assertion"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type Tcp_IPv int32

const (
	Tcp_DEFAULT Tcp_IPv = 0
	Tcp_IPV4    Tcp_IPv = 4
	Tcp_IPV6    Tcp_IPv = 6
	Tcp_BOTH    Tcp_IPv = 10
)

// Enum value maps for Tcp_IPv.
var (
	Tcp_IPv_name = map[int32]string{
		0:  "DEFAULT",
		4:  "IPV4",
		6:  "IPV6",
		10: "BOTH",
	}
	Tcp_IPv_value = map[string]int32{
		"DEFAULT": 0,
		"IPV4":    4,
		"IPV6":    6,
		"BOTH":    10,
	}
)

func (x Tcp_IPv) Enum() *Tcp_IPv {
	p := new(Tcp_IPv)
	*p = x
	return p
}

func (x Tcp_IPv) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (Tcp_IPv) Descriptor() protoreflect.EnumDescriptor {
	return file_tcp_proto_enumTypes[0].Descriptor()
}

func (Tcp_IPv) Type() protoreflect.EnumType {
	return &file_tcp_proto_enumTypes[0]
}

func (x Tcp_IPv) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use Tcp_IPv.Descriptor instead.
func (Tcp_IPv) EnumDescriptor() ([]byte, []int) {
	return file_tcp_proto_rawDescGZIP(), []int{1, 0}
}

type Probe struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Metadata   *probe.Metadata        `protobuf:"bytes,2,opt,name=Metadata,json=metadata,proto3" json:"Metadata,omitempty"`
	Assertions []*assertion.Assertion `protobuf:"bytes,74,rep,name=Assertions,json=assertions,proto3" json:"Assertions,omitempty"`
	Spec       *Tcp                   `protobuf:"bytes,100,opt,name=Spec,json=spec,proto3" json:"Spec,omitempty"`
}

func (x *Probe) Reset() {
	*x = Probe{}
	if protoimpl.UnsafeEnabled {
		mi := &file_tcp_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Probe) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Probe) ProtoMessage() {}

func (x *Probe) ProtoReflect() protoreflect.Message {
	mi := &file_tcp_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Probe.ProtoReflect.Descriptor instead.
func (*Probe) Descriptor() ([]byte, []int) {
	return file_tcp_proto_rawDescGZIP(), []int{0}
}

func (x *Probe) GetMetadata() *probe.Metadata {
	if x != nil {
		return x.Metadata
	}
	return nil
}

func (x *Probe) GetAssertions() []*assertion.Assertion {
	if x != nil {
		return x.Assertions
	}
	return nil
}

func (x *Probe) GetSpec() *Tcp {
	if x != nil {
		return x.Spec
	}
	return nil
}

type Tcp struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Host        string `protobuf:"bytes,10,opt,name=Host,json=host,proto3" json:"Host,omitempty"`
	Port        int32  `protobuf:"varint,15,opt,name=Port,json=port,proto3" json:"Port,omitempty"`
	PayloadSize int32  `protobuf:"varint,20,opt,name=PayloadSize,json=payload_size,proto3" json:"PayloadSize,omitempty"`
}

func (x *Tcp) Reset() {
	*x = Tcp{}
	if protoimpl.UnsafeEnabled {
		mi := &file_tcp_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Tcp) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Tcp) ProtoMessage() {}

func (x *Tcp) ProtoReflect() protoreflect.Message {
	mi := &file_tcp_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Tcp.ProtoReflect.Descriptor instead.
func (*Tcp) Descriptor() ([]byte, []int) {
	return file_tcp_proto_rawDescGZIP(), []int{1}
}

func (x *Tcp) GetHost() string {
	if x != nil {
		return x.Host
	}
	return ""
}

func (x *Tcp) GetPort() int32 {
	if x != nil {
		return x.Port
	}
	return 0
}

func (x *Tcp) GetPayloadSize() int32 {
	if x != nil {
		return x.PayloadSize
	}
	return 0
}

var File_tcp_proto protoreflect.FileDescriptor

var file_tcp_proto_rawDesc = []byte{
	0x0a, 0x09, 0x74, 0x63, 0x70, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x05, 0x70, 0x72, 0x6f,
	0x62, 0x65, 0x1a, 0x14, 0x70, 0x72, 0x6f, 0x62, 0x65, 0x5f, 0x6d, 0x65, 0x74, 0x61, 0x64, 0x61,
	0x74, 0x61, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x15, 0x70, 0x72, 0x6f, 0x62, 0x65, 0x5f,
	0x61, 0x73, 0x73, 0x65, 0x72, 0x74, 0x69, 0x6f, 0x6e, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22,
	0x86, 0x01, 0x0a, 0x05, 0x50, 0x72, 0x6f, 0x62, 0x65, 0x12, 0x2b, 0x0a, 0x08, 0x4d, 0x65, 0x74,
	0x61, 0x64, 0x61, 0x74, 0x61, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0f, 0x2e, 0x70, 0x72,
	0x6f, 0x62, 0x65, 0x2e, 0x4d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x52, 0x08, 0x6d, 0x65,
	0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x12, 0x30, 0x0a, 0x0a, 0x41, 0x73, 0x73, 0x65, 0x72, 0x74,
	0x69, 0x6f, 0x6e, 0x73, 0x18, 0x4a, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x10, 0x2e, 0x70, 0x72, 0x6f,
	0x62, 0x65, 0x2e, 0x41, 0x73, 0x73, 0x65, 0x72, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x0a, 0x61, 0x73,
	0x73, 0x65, 0x72, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x12, 0x1e, 0x0a, 0x04, 0x53, 0x70, 0x65, 0x63,
	0x18, 0x64, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0a, 0x2e, 0x70, 0x72, 0x6f, 0x62, 0x65, 0x2e, 0x74,
	0x63, 0x70, 0x52, 0x04, 0x73, 0x70, 0x65, 0x63, 0x22, 0x82, 0x01, 0x0a, 0x03, 0x74, 0x63, 0x70,
	0x12, 0x12, 0x0a, 0x04, 0x48, 0x6f, 0x73, 0x74, 0x18, 0x0a, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04,
	0x68, 0x6f, 0x73, 0x74, 0x12, 0x12, 0x0a, 0x04, 0x50, 0x6f, 0x72, 0x74, 0x18, 0x0f, 0x20, 0x01,
	0x28, 0x05, 0x52, 0x04, 0x70, 0x6f, 0x72, 0x74, 0x12, 0x21, 0x0a, 0x0b, 0x50, 0x61, 0x79, 0x6c,
	0x6f, 0x61, 0x64, 0x53, 0x69, 0x7a, 0x65, 0x18, 0x14, 0x20, 0x01, 0x28, 0x05, 0x52, 0x0c, 0x70,
	0x61, 0x79, 0x6c, 0x6f, 0x61, 0x64, 0x5f, 0x73, 0x69, 0x7a, 0x65, 0x22, 0x30, 0x0a, 0x03, 0x49,
	0x50, 0x76, 0x12, 0x0b, 0x0a, 0x07, 0x44, 0x45, 0x46, 0x41, 0x55, 0x4c, 0x54, 0x10, 0x00, 0x12,
	0x08, 0x0a, 0x04, 0x49, 0x50, 0x56, 0x34, 0x10, 0x04, 0x12, 0x08, 0x0a, 0x04, 0x49, 0x50, 0x56,
	0x36, 0x10, 0x06, 0x12, 0x08, 0x0a, 0x04, 0x42, 0x4f, 0x54, 0x48, 0x10, 0x0a, 0x42, 0x28, 0x5a,
	0x26, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x76, 0x69, 0x6e, 0x63,
	0x6f, 0x6c, 0x6c, 0x2f, 0x76, 0x69, 0x67, 0x69, 0x65, 0x2f, 0x70, 0x6b, 0x67, 0x2f, 0x70, 0x72,
	0x6f, 0x62, 0x65, 0x2f, 0x74, 0x63, 0x70, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_tcp_proto_rawDescOnce sync.Once
	file_tcp_proto_rawDescData = file_tcp_proto_rawDesc
)

func file_tcp_proto_rawDescGZIP() []byte {
	file_tcp_proto_rawDescOnce.Do(func() {
		file_tcp_proto_rawDescData = protoimpl.X.CompressGZIP(file_tcp_proto_rawDescData)
	})
	return file_tcp_proto_rawDescData
}

var file_tcp_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_tcp_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_tcp_proto_goTypes = []interface{}{
	(Tcp_IPv)(0),                // 0: probe.tcp.IPv
	(*Probe)(nil),               // 1: probe.Probe
	(*Tcp)(nil),                 // 2: probe.tcp
	(*probe.Metadata)(nil),      // 3: probe.Metadata
	(*assertion.Assertion)(nil), // 4: probe.Assertion
}
var file_tcp_proto_depIdxs = []int32{
	3, // 0: probe.Probe.Metadata:type_name -> probe.Metadata
	4, // 1: probe.Probe.Assertions:type_name -> probe.Assertion
	2, // 2: probe.Probe.Spec:type_name -> probe.tcp
	3, // [3:3] is the sub-list for method output_type
	3, // [3:3] is the sub-list for method input_type
	3, // [3:3] is the sub-list for extension type_name
	3, // [3:3] is the sub-list for extension extendee
	0, // [0:3] is the sub-list for field type_name
}

func init() { file_tcp_proto_init() }
func file_tcp_proto_init() {
	if File_tcp_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_tcp_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Probe); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_tcp_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Tcp); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_tcp_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_tcp_proto_goTypes,
		DependencyIndexes: file_tcp_proto_depIdxs,
		EnumInfos:         file_tcp_proto_enumTypes,
		MessageInfos:      file_tcp_proto_msgTypes,
	}.Build()
	File_tcp_proto = out.File
	file_tcp_proto_rawDesc = nil
	file_tcp_proto_goTypes = nil
	file_tcp_proto_depIdxs = nil
}
