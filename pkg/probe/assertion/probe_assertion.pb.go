// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.27.1
// 	protoc        v3.19.1
// source: probe_assertion.proto

package assertion

import (
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

type AssertionMethod int32

const (
	Assertion_EQUAL    AssertionMethod = 0
	Assertion_SUP      AssertionMethod = 4
	Assertion_LESS     AssertionMethod = 6
	Assertion_CONTAINS AssertionMethod = 10
	Assertion_REGEX    AssertionMethod = 21
)

// Enum value maps for AssertionMethod.
var (
	AssertionMethod_name = map[int32]string{
		0:  "EQUAL",
		4:  "SUP",
		6:  "LESS",
		10: "CONTAINS",
		21: "REGEX",
	}
	AssertionMethod_value = map[string]int32{
		"EQUAL":    0,
		"SUP":      4,
		"LESS":     6,
		"CONTAINS": 10,
		"REGEX":    21,
	}
)

func (x AssertionMethod) Enum() *AssertionMethod {
	p := new(AssertionMethod)
	*p = x
	return p
}

func (x AssertionMethod) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (AssertionMethod) Descriptor() protoreflect.EnumDescriptor {
	return file_probe_assertion_proto_enumTypes[0].Descriptor()
}

func (AssertionMethod) Type() protoreflect.EnumType {
	return &file_probe_assertion_proto_enumTypes[0]
}

func (x AssertionMethod) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use AssertionMethod.Descriptor instead.
func (AssertionMethod) EnumDescriptor() ([]byte, []int) {
	return file_probe_assertion_proto_rawDescGZIP(), []int{0, 0}
}

type AssertionResultResult int32

const (
	AssertionResult_UNDEFINED AssertionResultResult = 0
	AssertionResult_PASS      AssertionResultResult = 4
	AssertionResult_FAIL      AssertionResultResult = 6
)

// Enum value maps for AssertionResultResult.
var (
	AssertionResultResult_name = map[int32]string{
		0: "UNDEFINED",
		4: "PASS",
		6: "FAIL",
	}
	AssertionResultResult_value = map[string]int32{
		"UNDEFINED": 0,
		"PASS":      4,
		"FAIL":      6,
	}
)

func (x AssertionResultResult) Enum() *AssertionResultResult {
	p := new(AssertionResultResult)
	*p = x
	return p
}

func (x AssertionResultResult) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (AssertionResultResult) Descriptor() protoreflect.EnumDescriptor {
	return file_probe_assertion_proto_enumTypes[1].Descriptor()
}

func (AssertionResultResult) Type() protoreflect.EnumType {
	return &file_probe_assertion_proto_enumTypes[1]
}

func (x AssertionResultResult) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use AssertionResultResult.Descriptor instead.
func (AssertionResultResult) EnumDescriptor() ([]byte, []int) {
	return file_probe_assertion_proto_rawDescGZIP(), []int{1, 0}
}

type Assertion struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Key    string   `protobuf:"bytes,1,opt,name=key,proto3" json:"key,omitempty"`
	Value  string   `protobuf:"bytes,75,opt,name=Value,json=value,proto3" json:"Value,omitempty"`
	Values []string `protobuf:"bytes,74,rep,name=Values,json=values,proto3" json:"Values,omitempty"`
}

func (x *Assertion) Reset() {
	*x = Assertion{}
	if protoimpl.UnsafeEnabled {
		mi := &file_probe_assertion_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Assertion) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Assertion) ProtoMessage() {}

func (x *Assertion) ProtoReflect() protoreflect.Message {
	mi := &file_probe_assertion_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Assertion.ProtoReflect.Descriptor instead.
func (*Assertion) Descriptor() ([]byte, []int) {
	return file_probe_assertion_proto_rawDescGZIP(), []int{0}
}

func (x *Assertion) GetKey() string {
	if x != nil {
		return x.Key
	}
	return ""
}

func (x *Assertion) GetValue() string {
	if x != nil {
		return x.Value
	}
	return ""
}

func (x *Assertion) GetValues() []string {
	if x != nil {
		return x.Values
	}
	return nil
}

type AssertionResult struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Assertion string   `protobuf:"bytes,1,opt,name=assertion,proto3" json:"assertion,omitempty"`
	Value     string   `protobuf:"bytes,75,opt,name=Value,json=value,proto3" json:"Value,omitempty"`
	Values    []string `protobuf:"bytes,74,rep,name=Values,json=values,proto3" json:"Values,omitempty"`
}

func (x *AssertionResult) Reset() {
	*x = AssertionResult{}
	if protoimpl.UnsafeEnabled {
		mi := &file_probe_assertion_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *AssertionResult) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AssertionResult) ProtoMessage() {}

func (x *AssertionResult) ProtoReflect() protoreflect.Message {
	mi := &file_probe_assertion_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AssertionResult.ProtoReflect.Descriptor instead.
func (*AssertionResult) Descriptor() ([]byte, []int) {
	return file_probe_assertion_proto_rawDescGZIP(), []int{1}
}

func (x *AssertionResult) GetAssertion() string {
	if x != nil {
		return x.Assertion
	}
	return ""
}

func (x *AssertionResult) GetValue() string {
	if x != nil {
		return x.Value
	}
	return ""
}

func (x *AssertionResult) GetValues() []string {
	if x != nil {
		return x.Values
	}
	return nil
}

var File_probe_assertion_proto protoreflect.FileDescriptor

var file_probe_assertion_proto_rawDesc = []byte{
	0x0a, 0x15, 0x70, 0x72, 0x6f, 0x62, 0x65, 0x5f, 0x61, 0x73, 0x73, 0x65, 0x72, 0x74, 0x69, 0x6f,
	0x6e, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x05, 0x70, 0x72, 0x6f, 0x62, 0x65, 0x22, 0x8c,
	0x01, 0x0a, 0x09, 0x41, 0x73, 0x73, 0x65, 0x72, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x10, 0x0a, 0x03,
	0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x14,
	0x0a, 0x05, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x4b, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x76,
	0x61, 0x6c, 0x75, 0x65, 0x12, 0x16, 0x0a, 0x06, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x73, 0x18, 0x4a,
	0x20, 0x03, 0x28, 0x09, 0x52, 0x06, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x73, 0x22, 0x3f, 0x0a, 0x06,
	0x6d, 0x65, 0x74, 0x68, 0x6f, 0x64, 0x12, 0x09, 0x0a, 0x05, 0x45, 0x51, 0x55, 0x41, 0x4c, 0x10,
	0x00, 0x12, 0x07, 0x0a, 0x03, 0x53, 0x55, 0x50, 0x10, 0x04, 0x12, 0x08, 0x0a, 0x04, 0x4c, 0x45,
	0x53, 0x53, 0x10, 0x06, 0x12, 0x0c, 0x0a, 0x08, 0x43, 0x4f, 0x4e, 0x54, 0x41, 0x49, 0x4e, 0x53,
	0x10, 0x0a, 0x12, 0x09, 0x0a, 0x05, 0x52, 0x45, 0x47, 0x45, 0x58, 0x10, 0x15, 0x22, 0x8a, 0x01,
	0x0a, 0x0f, 0x41, 0x73, 0x73, 0x65, 0x72, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x65, 0x73, 0x75, 0x6c,
	0x74, 0x12, 0x1c, 0x0a, 0x09, 0x61, 0x73, 0x73, 0x65, 0x72, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x61, 0x73, 0x73, 0x65, 0x72, 0x74, 0x69, 0x6f, 0x6e, 0x12,
	0x14, 0x0a, 0x05, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x4b, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05,
	0x76, 0x61, 0x6c, 0x75, 0x65, 0x12, 0x16, 0x0a, 0x06, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x73, 0x18,
	0x4a, 0x20, 0x03, 0x28, 0x09, 0x52, 0x06, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x73, 0x22, 0x2b, 0x0a,
	0x06, 0x72, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x12, 0x0d, 0x0a, 0x09, 0x55, 0x4e, 0x44, 0x45, 0x46,
	0x49, 0x4e, 0x45, 0x44, 0x10, 0x00, 0x12, 0x08, 0x0a, 0x04, 0x50, 0x41, 0x53, 0x53, 0x10, 0x04,
	0x12, 0x08, 0x0a, 0x04, 0x46, 0x41, 0x49, 0x4c, 0x10, 0x06, 0x42, 0x2e, 0x5a, 0x2c, 0x67, 0x69,
	0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x76, 0x69, 0x6e, 0x63, 0x6f, 0x6c, 0x6c,
	0x2f, 0x76, 0x69, 0x67, 0x69, 0x65, 0x2f, 0x70, 0x6b, 0x67, 0x2f, 0x70, 0x72, 0x6f, 0x62, 0x65,
	0x2f, 0x61, 0x73, 0x73, 0x65, 0x72, 0x74, 0x69, 0x6f, 0x6e, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x33,
}

var (
	file_probe_assertion_proto_rawDescOnce sync.Once
	file_probe_assertion_proto_rawDescData = file_probe_assertion_proto_rawDesc
)

func file_probe_assertion_proto_rawDescGZIP() []byte {
	file_probe_assertion_proto_rawDescOnce.Do(func() {
		file_probe_assertion_proto_rawDescData = protoimpl.X.CompressGZIP(file_probe_assertion_proto_rawDescData)
	})
	return file_probe_assertion_proto_rawDescData
}

var file_probe_assertion_proto_enumTypes = make([]protoimpl.EnumInfo, 2)
var file_probe_assertion_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_probe_assertion_proto_goTypes = []interface{}{
	(AssertionMethod)(0),       // 0: probe.Assertion.method
	(AssertionResultResult)(0), // 1: probe.AssertionResult.result
	(*Assertion)(nil),          // 2: probe.Assertion
	(*AssertionResult)(nil),    // 3: probe.AssertionResult
}
var file_probe_assertion_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_probe_assertion_proto_init() }
func file_probe_assertion_proto_init() {
	if File_probe_assertion_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_probe_assertion_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Assertion); i {
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
		file_probe_assertion_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*AssertionResult); i {
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
			RawDescriptor: file_probe_assertion_proto_rawDesc,
			NumEnums:      2,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_probe_assertion_proto_goTypes,
		DependencyIndexes: file_probe_assertion_proto_depIdxs,
		EnumInfos:         file_probe_assertion_proto_enumTypes,
		MessageInfos:      file_probe_assertion_proto_msgTypes,
	}.Build()
	File_probe_assertion_proto = out.File
	file_probe_assertion_proto_rawDesc = nil
	file_probe_assertion_proto_goTypes = nil
	file_probe_assertion_proto_depIdxs = nil
}
