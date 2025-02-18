// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.36.5
// 	protoc        v4.23.3
// source: relay.proto

package api

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
	unsafe "unsafe"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// Define the Packet message with required fields.
type Packet struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Source        int64                  `protobuf:"varint,1,opt,name=source,proto3" json:"source,omitempty"`
	Destination   int64                  `protobuf:"varint,2,opt,name=destination,proto3" json:"destination,omitempty"`
	Content       []byte                 `protobuf:"bytes,3,opt,name=content,proto3" json:"content,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *Packet) Reset() {
	*x = Packet{}
	mi := &file_relay_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Packet) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Packet) ProtoMessage() {}

func (x *Packet) ProtoReflect() protoreflect.Message {
	mi := &file_relay_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Packet.ProtoReflect.Descriptor instead.
func (*Packet) Descriptor() ([]byte, []int) {
	return file_relay_proto_rawDescGZIP(), []int{0}
}

func (x *Packet) GetSource() int64 {
	if x != nil {
		return x.Source
	}
	return 0
}

func (x *Packet) GetDestination() int64 {
	if x != nil {
		return x.Destination
	}
	return 0
}

func (x *Packet) GetContent() []byte {
	if x != nil {
		return x.Content
	}
	return nil
}

type Peer struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	VirtualIp     string                 `protobuf:"bytes,1,opt,name=virtual_ip,json=virtualIp,proto3" json:"virtual_ip,omitempty"`
	PeerId        int64                  `protobuf:"varint,2,opt,name=peer_id,json=peerId,proto3" json:"peer_id,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *Peer) Reset() {
	*x = Peer{}
	mi := &file_relay_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Peer) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Peer) ProtoMessage() {}

func (x *Peer) ProtoReflect() protoreflect.Message {
	mi := &file_relay_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Peer.ProtoReflect.Descriptor instead.
func (*Peer) Descriptor() ([]byte, []int) {
	return file_relay_proto_rawDescGZIP(), []int{1}
}

func (x *Peer) GetVirtualIp() string {
	if x != nil {
		return x.VirtualIp
	}
	return ""
}

func (x *Peer) GetPeerId() int64 {
	if x != nil {
		return x.PeerId
	}
	return 0
}

type PeerList struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Peers         []*Peer                `protobuf:"bytes,1,rep,name=peers,proto3" json:"peers,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *PeerList) Reset() {
	*x = PeerList{}
	mi := &file_relay_proto_msgTypes[2]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *PeerList) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PeerList) ProtoMessage() {}

func (x *PeerList) ProtoReflect() protoreflect.Message {
	mi := &file_relay_proto_msgTypes[2]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PeerList.ProtoReflect.Descriptor instead.
func (*PeerList) Descriptor() ([]byte, []int) {
	return file_relay_proto_rawDescGZIP(), []int{2}
}

func (x *PeerList) GetPeers() []*Peer {
	if x != nil {
		return x.Peers
	}
	return nil
}

var File_relay_proto protoreflect.FileDescriptor

var file_relay_proto_rawDesc = string([]byte{
	0x0a, 0x0b, 0x72, 0x65, 0x6c, 0x61, 0x79, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x03, 0x61,
	0x70, 0x69, 0x22, 0x5c, 0x0a, 0x06, 0x50, 0x61, 0x63, 0x6b, 0x65, 0x74, 0x12, 0x16, 0x0a, 0x06,
	0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x03, 0x52, 0x06, 0x73, 0x6f,
	0x75, 0x72, 0x63, 0x65, 0x12, 0x20, 0x0a, 0x0b, 0x64, 0x65, 0x73, 0x74, 0x69, 0x6e, 0x61, 0x74,
	0x69, 0x6f, 0x6e, 0x18, 0x02, 0x20, 0x01, 0x28, 0x03, 0x52, 0x0b, 0x64, 0x65, 0x73, 0x74, 0x69,
	0x6e, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x18, 0x0a, 0x07, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e,
	0x74, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x07, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74,
	0x22, 0x3e, 0x0a, 0x04, 0x50, 0x65, 0x65, 0x72, 0x12, 0x1d, 0x0a, 0x0a, 0x76, 0x69, 0x72, 0x74,
	0x75, 0x61, 0x6c, 0x5f, 0x69, 0x70, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x76, 0x69,
	0x72, 0x74, 0x75, 0x61, 0x6c, 0x49, 0x70, 0x12, 0x17, 0x0a, 0x07, 0x70, 0x65, 0x65, 0x72, 0x5f,
	0x69, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x03, 0x52, 0x06, 0x70, 0x65, 0x65, 0x72, 0x49, 0x64,
	0x22, 0x2b, 0x0a, 0x08, 0x50, 0x65, 0x65, 0x72, 0x4c, 0x69, 0x73, 0x74, 0x12, 0x1f, 0x0a, 0x05,
	0x70, 0x65, 0x65, 0x72, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x09, 0x2e, 0x61, 0x70,
	0x69, 0x2e, 0x50, 0x65, 0x65, 0x72, 0x52, 0x05, 0x70, 0x65, 0x65, 0x72, 0x73, 0x32, 0x58, 0x0a,
	0x05, 0x52, 0x65, 0x6c, 0x61, 0x79, 0x12, 0x28, 0x0a, 0x0c, 0x52, 0x65, 0x67, 0x69, 0x73, 0x74,
	0x65, 0x72, 0x50, 0x65, 0x65, 0x72, 0x12, 0x09, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x50, 0x65, 0x65,
	0x72, 0x1a, 0x0d, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x50, 0x65, 0x65, 0x72, 0x4c, 0x69, 0x73, 0x74,
	0x12, 0x25, 0x0a, 0x05, 0x50, 0x72, 0x6f, 0x78, 0x79, 0x12, 0x0b, 0x2e, 0x61, 0x70, 0x69, 0x2e,
	0x50, 0x61, 0x63, 0x6b, 0x65, 0x74, 0x1a, 0x0b, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x50, 0x61, 0x63,
	0x6b, 0x65, 0x74, 0x28, 0x01, 0x30, 0x01, 0x42, 0x07, 0x5a, 0x05, 0x2e, 0x2f, 0x61, 0x70, 0x69,
	0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
})

var (
	file_relay_proto_rawDescOnce sync.Once
	file_relay_proto_rawDescData []byte
)

func file_relay_proto_rawDescGZIP() []byte {
	file_relay_proto_rawDescOnce.Do(func() {
		file_relay_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_relay_proto_rawDesc), len(file_relay_proto_rawDesc)))
	})
	return file_relay_proto_rawDescData
}

var file_relay_proto_msgTypes = make([]protoimpl.MessageInfo, 3)
var file_relay_proto_goTypes = []any{
	(*Packet)(nil),   // 0: api.Packet
	(*Peer)(nil),     // 1: api.Peer
	(*PeerList)(nil), // 2: api.PeerList
}
var file_relay_proto_depIdxs = []int32{
	1, // 0: api.PeerList.peers:type_name -> api.Peer
	1, // 1: api.Relay.RegisterPeer:input_type -> api.Peer
	0, // 2: api.Relay.Proxy:input_type -> api.Packet
	2, // 3: api.Relay.RegisterPeer:output_type -> api.PeerList
	0, // 4: api.Relay.Proxy:output_type -> api.Packet
	3, // [3:5] is the sub-list for method output_type
	1, // [1:3] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_relay_proto_init() }
func file_relay_proto_init() {
	if File_relay_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_relay_proto_rawDesc), len(file_relay_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   3,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_relay_proto_goTypes,
		DependencyIndexes: file_relay_proto_depIdxs,
		MessageInfos:      file_relay_proto_msgTypes,
	}.Build()
	File_relay_proto = out.File
	file_relay_proto_goTypes = nil
	file_relay_proto_depIdxs = nil
}
