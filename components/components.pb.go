// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.26.0-devel
// 	protoc        v3.15.2
// source: components.proto

package components

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

type Position struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	X int64 `protobuf:"varint,1,opt,name=x,proto3" json:"x,omitempty"`
	Y int64 `protobuf:"varint,2,opt,name=y,proto3" json:"y,omitempty"`
}

func (x *Position) Reset() {
	*x = Position{}
	if protoimpl.UnsafeEnabled {
		mi := &file_components_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Position) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Position) ProtoMessage() {}

func (x *Position) ProtoReflect() protoreflect.Message {
	mi := &file_components_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Position.ProtoReflect.Descriptor instead.
func (*Position) Descriptor() ([]byte, []int) {
	return file_components_proto_rawDescGZIP(), []int{0}
}

func (x *Position) GetX() int64 {
	if x != nil {
		return x.X
	}
	return 0
}

func (x *Position) GetY() int64 {
	if x != nil {
		return x.Y
	}
	return 0
}

type Moveable struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// How many units per tick it moves.
	VelX int64 `protobuf:"varint,1,opt,name=vel_x,json=velX,proto3" json:"vel_x,omitempty"`
	// How many units per tick it moves.
	VelY int64 `protobuf:"varint,2,opt,name=vel_y,json=velY,proto3" json:"vel_y,omitempty"`
}

func (x *Moveable) Reset() {
	*x = Moveable{}
	if protoimpl.UnsafeEnabled {
		mi := &file_components_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Moveable) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Moveable) ProtoMessage() {}

func (x *Moveable) ProtoReflect() protoreflect.Message {
	mi := &file_components_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Moveable.ProtoReflect.Descriptor instead.
func (*Moveable) Descriptor() ([]byte, []int) {
	return file_components_proto_rawDescGZIP(), []int{1}
}

func (x *Moveable) GetVelX() int64 {
	if x != nil {
		return x.VelX
	}
	return 0
}

func (x *Moveable) GetVelY() int64 {
	if x != nil {
		return x.VelY
	}
	return 0
}

type Named struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Name string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
}

func (x *Named) Reset() {
	*x = Named{}
	if protoimpl.UnsafeEnabled {
		mi := &file_components_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Named) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Named) ProtoMessage() {}

func (x *Named) ProtoReflect() protoreflect.Message {
	mi := &file_components_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Named.ProtoReflect.Descriptor instead.
func (*Named) Descriptor() ([]byte, []int) {
	return file_components_proto_rawDescGZIP(), []int{2}
}

func (x *Named) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

type Looker struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *Looker) Reset() {
	*x = Looker{}
	if protoimpl.UnsafeEnabled {
		mi := &file_components_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Looker) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Looker) ProtoMessage() {}

func (x *Looker) ProtoReflect() protoreflect.Message {
	mi := &file_components_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Looker.ProtoReflect.Descriptor instead.
func (*Looker) Descriptor() ([]byte, []int) {
	return file_components_proto_rawDescGZIP(), []int{3}
}

type Speaker struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Range float32 `protobuf:"fixed32,1,opt,name=range,proto3" json:"range,omitempty"`
}

func (x *Speaker) Reset() {
	*x = Speaker{}
	if protoimpl.UnsafeEnabled {
		mi := &file_components_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Speaker) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Speaker) ProtoMessage() {}

func (x *Speaker) ProtoReflect() protoreflect.Message {
	mi := &file_components_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Speaker.ProtoReflect.Descriptor instead.
func (*Speaker) Descriptor() ([]byte, []int) {
	return file_components_proto_rawDescGZIP(), []int{4}
}

func (x *Speaker) GetRange() float32 {
	if x != nil {
		return x.Range
	}
	return 0
}

type Render struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Char  string `protobuf:"bytes,1,opt,name=char,proto3" json:"char,omitempty"`
	Color uint32 `protobuf:"varint,2,opt,name=color,proto3" json:"color,omitempty"`
}

func (x *Render) Reset() {
	*x = Render{}
	if protoimpl.UnsafeEnabled {
		mi := &file_components_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Render) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Render) ProtoMessage() {}

func (x *Render) ProtoReflect() protoreflect.Message {
	mi := &file_components_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Render.ProtoReflect.Descriptor instead.
func (*Render) Descriptor() ([]byte, []int) {
	return file_components_proto_rawDescGZIP(), []int{5}
}

func (x *Render) GetChar() string {
	if x != nil {
		return x.Char
	}
	return ""
}

func (x *Render) GetColor() uint32 {
	if x != nil {
		return x.Color
	}
	return 0
}

type Readable struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Text string `protobuf:"bytes,1,opt,name=text,proto3" json:"text,omitempty"`
}

func (x *Readable) Reset() {
	*x = Readable{}
	if protoimpl.UnsafeEnabled {
		mi := &file_components_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Readable) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Readable) ProtoMessage() {}

func (x *Readable) ProtoReflect() protoreflect.Message {
	mi := &file_components_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Readable.ProtoReflect.Descriptor instead.
func (*Readable) Descriptor() ([]byte, []int) {
	return file_components_proto_rawDescGZIP(), []int{6}
}

func (x *Readable) GetText() string {
	if x != nil {
		return x.Text
	}
	return ""
}

var File_components_proto protoreflect.FileDescriptor

var file_components_proto_rawDesc = []byte{
	0x0a, 0x10, 0x63, 0x6f, 0x6d, 0x70, 0x6f, 0x6e, 0x65, 0x6e, 0x74, 0x73, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x12, 0x0a, 0x63, 0x6f, 0x6d, 0x70, 0x6f, 0x6e, 0x65, 0x6e, 0x74, 0x73, 0x22, 0x26,
	0x0a, 0x08, 0x50, 0x6f, 0x73, 0x69, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x0c, 0x0a, 0x01, 0x78, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x03, 0x52, 0x01, 0x78, 0x12, 0x0c, 0x0a, 0x01, 0x79, 0x18, 0x02, 0x20,
	0x01, 0x28, 0x03, 0x52, 0x01, 0x79, 0x22, 0x34, 0x0a, 0x08, 0x4d, 0x6f, 0x76, 0x65, 0x61, 0x62,
	0x6c, 0x65, 0x12, 0x13, 0x0a, 0x05, 0x76, 0x65, 0x6c, 0x5f, 0x78, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x03, 0x52, 0x04, 0x76, 0x65, 0x6c, 0x58, 0x12, 0x13, 0x0a, 0x05, 0x76, 0x65, 0x6c, 0x5f, 0x79,
	0x18, 0x02, 0x20, 0x01, 0x28, 0x03, 0x52, 0x04, 0x76, 0x65, 0x6c, 0x59, 0x22, 0x1b, 0x0a, 0x05,
	0x4e, 0x61, 0x6d, 0x65, 0x64, 0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x22, 0x08, 0x0a, 0x06, 0x4c, 0x6f, 0x6f,
	0x6b, 0x65, 0x72, 0x22, 0x1f, 0x0a, 0x07, 0x53, 0x70, 0x65, 0x61, 0x6b, 0x65, 0x72, 0x12, 0x14,
	0x0a, 0x05, 0x72, 0x61, 0x6e, 0x67, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x02, 0x52, 0x05, 0x72,
	0x61, 0x6e, 0x67, 0x65, 0x22, 0x32, 0x0a, 0x06, 0x52, 0x65, 0x6e, 0x64, 0x65, 0x72, 0x12, 0x12,
	0x0a, 0x04, 0x63, 0x68, 0x61, 0x72, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x63, 0x68,
	0x61, 0x72, 0x12, 0x14, 0x0a, 0x05, 0x63, 0x6f, 0x6c, 0x6f, 0x72, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x0d, 0x52, 0x05, 0x63, 0x6f, 0x6c, 0x6f, 0x72, 0x22, 0x1e, 0x0a, 0x08, 0x52, 0x65, 0x61, 0x64,
	0x61, 0x62, 0x6c, 0x65, 0x12, 0x12, 0x0a, 0x04, 0x74, 0x65, 0x78, 0x74, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x04, 0x74, 0x65, 0x78, 0x74, 0x42, 0x27, 0x5a, 0x25, 0x67, 0x69, 0x74, 0x68,
	0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x63, 0x6f, 0x64, 0x65, 0x2d, 0x63, 0x65, 0x6c, 0x6c,
	0x2f, 0x65, 0x73, 0x69, 0x76, 0x65, 0x2f, 0x63, 0x6f, 0x6d, 0x70, 0x6f, 0x6e, 0x65, 0x6e, 0x74,
	0x73, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_components_proto_rawDescOnce sync.Once
	file_components_proto_rawDescData = file_components_proto_rawDesc
)

func file_components_proto_rawDescGZIP() []byte {
	file_components_proto_rawDescOnce.Do(func() {
		file_components_proto_rawDescData = protoimpl.X.CompressGZIP(file_components_proto_rawDescData)
	})
	return file_components_proto_rawDescData
}

var file_components_proto_msgTypes = make([]protoimpl.MessageInfo, 7)
var file_components_proto_goTypes = []interface{}{
	(*Position)(nil), // 0: components.Position
	(*Moveable)(nil), // 1: components.Moveable
	(*Named)(nil),    // 2: components.Named
	(*Looker)(nil),   // 3: components.Looker
	(*Speaker)(nil),  // 4: components.Speaker
	(*Render)(nil),   // 5: components.Render
	(*Readable)(nil), // 6: components.Readable
}
var file_components_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_components_proto_init() }
func file_components_proto_init() {
	if File_components_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_components_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Position); i {
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
		file_components_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Moveable); i {
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
		file_components_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Named); i {
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
		file_components_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Looker); i {
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
		file_components_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Speaker); i {
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
		file_components_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Render); i {
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
		file_components_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Readable); i {
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
			RawDescriptor: file_components_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   7,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_components_proto_goTypes,
		DependencyIndexes: file_components_proto_depIdxs,
		MessageInfos:      file_components_proto_msgTypes,
	}.Build()
	File_components_proto = out.File
	file_components_proto_rawDesc = nil
	file_components_proto_goTypes = nil
	file_components_proto_depIdxs = nil
}
