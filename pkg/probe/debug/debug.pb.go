// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.0
// 	protoc        v4.25.0
// source: debug.proto

package debug

import (
	probe "github.com/vincoll/vigie/pkg/probe"
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

type Debug struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Name     string          `protobuf:"bytes,1,opt,name=Name,json=name,proto3" json:"Name,omitempty"`
	Test     string          `protobuf:"bytes,2,opt,name=Test,json=test,proto3" json:"Test,omitempty"`
	Metadata *probe.Metadata `protobuf:"bytes,90,opt,name=Metadata,json=metadata,proto3" json:"Metadata,omitempty"`
}

func (x *Debug) Reset() {
	*x = Debug{}
	if protoimpl.UnsafeEnabled {
		mi := &file_debug_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Debug) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Debug) ProtoMessage() {}

func (x *Debug) ProtoReflect() protoreflect.Message {
	mi := &file_debug_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Debug.ProtoReflect.Descriptor instead.
func (*Debug) Descriptor() ([]byte, []int) {
	return file_debug_proto_rawDescGZIP(), []int{0}
}

func (x *Debug) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *Debug) GetTest() string {
	if x != nil {
		return x.Test
	}
	return ""
}

func (x *Debug) GetMetadata() *probe.Metadata {
	if x != nil {
		return x.Metadata
	}
	return nil
}

var File_debug_proto protoreflect.FileDescriptor

var file_debug_proto_rawDesc = []byte{
	0x0a, 0x0b, 0x64, 0x65, 0x62, 0x75, 0x67, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x05, 0x70,
	0x72, 0x6f, 0x62, 0x65, 0x1a, 0x14, 0x70, 0x72, 0x6f, 0x62, 0x65, 0x5f, 0x6d, 0x65, 0x74, 0x61,
	0x64, 0x61, 0x74, 0x61, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x5c, 0x0a, 0x05, 0x44, 0x65,
	0x62, 0x75, 0x67, 0x12, 0x12, 0x0a, 0x04, 0x4e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x12, 0x0a, 0x04, 0x54, 0x65, 0x73, 0x74, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x74, 0x65, 0x73, 0x74, 0x12, 0x2b, 0x0a, 0x08, 0x4d,
	0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x18, 0x5a, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0f, 0x2e,
	0x70, 0x72, 0x6f, 0x62, 0x65, 0x2e, 0x4d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x52, 0x08,
	0x6d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x42, 0x2a, 0x5a, 0x28, 0x67, 0x69, 0x74, 0x68,
	0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x76, 0x69, 0x6e, 0x63, 0x6f, 0x6c, 0x6c, 0x2f, 0x76,
	0x69, 0x67, 0x69, 0x65, 0x2f, 0x70, 0x6b, 0x67, 0x2f, 0x70, 0x72, 0x6f, 0x62, 0x65, 0x2f, 0x64,
	0x65, 0x62, 0x75, 0x67, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_debug_proto_rawDescOnce sync.Once
	file_debug_proto_rawDescData = file_debug_proto_rawDesc
)

func file_debug_proto_rawDescGZIP() []byte {
	file_debug_proto_rawDescOnce.Do(func() {
		file_debug_proto_rawDescData = protoimpl.X.CompressGZIP(file_debug_proto_rawDescData)
	})
	return file_debug_proto_rawDescData
}

var file_debug_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_debug_proto_goTypes = []interface{}{
	(*Debug)(nil),          // 0: probe.Debug
	(*probe.Metadata)(nil), // 1: probe.Metadata
}
var file_debug_proto_depIdxs = []int32{
	1, // 0: probe.Debug.Metadata:type_name -> probe.Metadata
	1, // [1:1] is the sub-list for method output_type
	1, // [1:1] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_debug_proto_init() }
func file_debug_proto_init() {
	if File_debug_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_debug_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Debug); i {
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
			RawDescriptor: file_debug_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_debug_proto_goTypes,
		DependencyIndexes: file_debug_proto_depIdxs,
		MessageInfos:      file_debug_proto_msgTypes,
	}.Build()
	File_debug_proto = out.File
	file_debug_proto_rawDesc = nil
	file_debug_proto_goTypes = nil
	file_debug_proto_depIdxs = nil
}
