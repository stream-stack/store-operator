// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.27.1
// 	protoc        v3.17.3
// source: xds.proto

package proto

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

type Partition struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Begin uint64    `protobuf:"varint,1,opt,name=begin,proto3" json:"begin,omitempty"`
	Store *StoreSet `protobuf:"bytes,2,opt,name=store,proto3" json:"store,omitempty"`
}

func (x *Partition) Reset() {
	*x = Partition{}
	if protoimpl.UnsafeEnabled {
		mi := &file_xds_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Partition) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Partition) ProtoMessage() {}

func (x *Partition) ProtoReflect() protoreflect.Message {
	mi := &file_xds_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Partition.ProtoReflect.Descriptor instead.
func (*Partition) Descriptor() ([]byte, []int) {
	return file_xds_proto_rawDescGZIP(), []int{0}
}

func (x *Partition) GetBegin() uint64 {
	if x != nil {
		return x.Begin
	}
	return 0
}

func (x *Partition) GetStore() *StoreSet {
	if x != nil {
		return x.Store
	}
	return nil
}

type Subscribe struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Name string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	Uri  string `protobuf:"bytes,2,opt,name=uri,proto3" json:"uri,omitempty"`
}

func (x *Subscribe) Reset() {
	*x = Subscribe{}
	if protoimpl.UnsafeEnabled {
		mi := &file_xds_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Subscribe) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Subscribe) ProtoMessage() {}

func (x *Subscribe) ProtoReflect() protoreflect.Message {
	mi := &file_xds_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Subscribe.ProtoReflect.Descriptor instead.
func (*Subscribe) Descriptor() ([]byte, []int) {
	return file_xds_proto_rawDescGZIP(), []int{1}
}

func (x *Subscribe) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *Subscribe) GetUri() string {
	if x != nil {
		return x.Uri
	}
	return ""
}

type StoreSet struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Name      string   `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	Namespace string   `protobuf:"bytes,2,opt,name=Namespace,proto3" json:"Namespace,omitempty"`
	Uris      []string `protobuf:"bytes,3,rep,name=uris,proto3" json:"uris,omitempty"`
}

func (x *StoreSet) Reset() {
	*x = StoreSet{}
	if protoimpl.UnsafeEnabled {
		mi := &file_xds_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *StoreSet) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*StoreSet) ProtoMessage() {}

func (x *StoreSet) ProtoReflect() protoreflect.Message {
	mi := &file_xds_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use StoreSet.ProtoReflect.Descriptor instead.
func (*StoreSet) Descriptor() ([]byte, []int) {
	return file_xds_proto_rawDescGZIP(), []int{2}
}

func (x *StoreSet) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *StoreSet) GetNamespace() string {
	if x != nil {
		return x.Namespace
	}
	return ""
}

func (x *StoreSet) GetUris() []string {
	if x != nil {
		return x.Uris
	}
	return nil
}

var File_xds_proto protoreflect.FileDescriptor

var file_xds_proto_rawDesc = []byte{
	0x0a, 0x09, 0x78, 0x64, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x42, 0x0a, 0x09, 0x50,
	0x61, 0x72, 0x74, 0x69, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x14, 0x0a, 0x05, 0x62, 0x65, 0x67, 0x69,
	0x6e, 0x18, 0x01, 0x20, 0x01, 0x28, 0x04, 0x52, 0x05, 0x62, 0x65, 0x67, 0x69, 0x6e, 0x12, 0x1f,
	0x0a, 0x05, 0x73, 0x74, 0x6f, 0x72, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x09, 0x2e,
	0x53, 0x74, 0x6f, 0x72, 0x65, 0x53, 0x65, 0x74, 0x52, 0x05, 0x73, 0x74, 0x6f, 0x72, 0x65, 0x22,
	0x31, 0x0a, 0x09, 0x53, 0x75, 0x62, 0x73, 0x63, 0x72, 0x69, 0x62, 0x65, 0x12, 0x12, 0x0a, 0x04,
	0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65,
	0x12, 0x10, 0x0a, 0x03, 0x75, 0x72, 0x69, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x75,
	0x72, 0x69, 0x22, 0x50, 0x0a, 0x08, 0x53, 0x74, 0x6f, 0x72, 0x65, 0x53, 0x65, 0x74, 0x12, 0x12,
	0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61,
	0x6d, 0x65, 0x12, 0x1c, 0x0a, 0x09, 0x4e, 0x61, 0x6d, 0x65, 0x73, 0x70, 0x61, 0x63, 0x65, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x4e, 0x61, 0x6d, 0x65, 0x73, 0x70, 0x61, 0x63, 0x65,
	0x12, 0x12, 0x0a, 0x04, 0x75, 0x72, 0x69, 0x73, 0x18, 0x03, 0x20, 0x03, 0x28, 0x09, 0x52, 0x04,
	0x75, 0x72, 0x69, 0x73, 0x42, 0x0a, 0x5a, 0x08, 0x2e, 0x2e, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_xds_proto_rawDescOnce sync.Once
	file_xds_proto_rawDescData = file_xds_proto_rawDesc
)

func file_xds_proto_rawDescGZIP() []byte {
	file_xds_proto_rawDescOnce.Do(func() {
		file_xds_proto_rawDescData = protoimpl.X.CompressGZIP(file_xds_proto_rawDescData)
	})
	return file_xds_proto_rawDescData
}

var file_xds_proto_msgTypes = make([]protoimpl.MessageInfo, 3)
var file_xds_proto_goTypes = []interface{}{
	(*Partition)(nil), // 0: Partition
	(*Subscribe)(nil), // 1: Subscribe
	(*StoreSet)(nil),  // 2: StoreSet
}
var file_xds_proto_depIdxs = []int32{
	2, // 0: Partition.store:type_name -> StoreSet
	1, // [1:1] is the sub-list for method output_type
	1, // [1:1] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_xds_proto_init() }
func file_xds_proto_init() {
	if File_xds_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_xds_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Partition); i {
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
		file_xds_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Subscribe); i {
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
		file_xds_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*StoreSet); i {
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
			RawDescriptor: file_xds_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   3,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_xds_proto_goTypes,
		DependencyIndexes: file_xds_proto_depIdxs,
		MessageInfos:      file_xds_proto_msgTypes,
	}.Build()
	File_xds_proto = out.File
	file_xds_proto_rawDesc = nil
	file_xds_proto_goTypes = nil
	file_xds_proto_depIdxs = nil
}
