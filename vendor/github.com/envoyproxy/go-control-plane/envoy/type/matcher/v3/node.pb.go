// Code generated by protoc-gen-go. DO NOT EDIT.
// source: envoy/type/matcher/v3/node.proto

package envoy_type_matcher_v3

import (
	fmt "fmt"
	_ "github.com/cncf/udpa/go/udpa/annotations"
	proto "github.com/golang/protobuf/proto"
	math "math"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion3 // please upgrade the proto package

type NodeMatcher struct {
	NodeId               *StringMatcher   `protobuf:"bytes,1,opt,name=node_id,json=nodeId,proto3" json:"node_id,omitempty"`
	NodeMetadatas        []*StructMatcher `protobuf:"bytes,2,rep,name=node_metadatas,json=nodeMetadatas,proto3" json:"node_metadatas,omitempty"`
	XXX_NoUnkeyedLiteral struct{}         `json:"-"`
	XXX_unrecognized     []byte           `json:"-"`
	XXX_sizecache        int32            `json:"-"`
}

func (m *NodeMatcher) Reset()         { *m = NodeMatcher{} }
func (m *NodeMatcher) String() string { return proto.CompactTextString(m) }
func (*NodeMatcher) ProtoMessage()    {}
func (*NodeMatcher) Descriptor() ([]byte, []int) {
	return fileDescriptor_ccc0c3eef80eb4fc, []int{0}
}

func (m *NodeMatcher) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_NodeMatcher.Unmarshal(m, b)
}
func (m *NodeMatcher) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_NodeMatcher.Marshal(b, m, deterministic)
}
func (m *NodeMatcher) XXX_Merge(src proto.Message) {
	xxx_messageInfo_NodeMatcher.Merge(m, src)
}
func (m *NodeMatcher) XXX_Size() int {
	return xxx_messageInfo_NodeMatcher.Size(m)
}
func (m *NodeMatcher) XXX_DiscardUnknown() {
	xxx_messageInfo_NodeMatcher.DiscardUnknown(m)
}

var xxx_messageInfo_NodeMatcher proto.InternalMessageInfo

func (m *NodeMatcher) GetNodeId() *StringMatcher {
	if m != nil {
		return m.NodeId
	}
	return nil
}

func (m *NodeMatcher) GetNodeMetadatas() []*StructMatcher {
	if m != nil {
		return m.NodeMetadatas
	}
	return nil
}

func init() {
	proto.RegisterType((*NodeMatcher)(nil), "envoy.type.matcher.v3.NodeMatcher")
}

func init() { proto.RegisterFile("envoy/type/matcher/v3/node.proto", fileDescriptor_ccc0c3eef80eb4fc) }

var fileDescriptor_ccc0c3eef80eb4fc = []byte{
	// 239 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0x52, 0x48, 0xcd, 0x2b, 0xcb,
	0xaf, 0xd4, 0x2f, 0xa9, 0x2c, 0x48, 0xd5, 0xcf, 0x4d, 0x2c, 0x49, 0xce, 0x48, 0x2d, 0xd2, 0x2f,
	0x33, 0xd6, 0xcf, 0xcb, 0x4f, 0x49, 0xd5, 0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x17, 0x12, 0x05, 0xab,
	0xd0, 0x03, 0xa9, 0xd0, 0x83, 0xaa, 0xd0, 0x2b, 0x33, 0x96, 0x52, 0xc2, 0xae, 0xb1, 0xb8, 0xa4,
	0x28, 0x33, 0x2f, 0x1d, 0xa2, 0x15, 0x8f, 0x9a, 0xd2, 0xe4, 0x12, 0xa8, 0x1a, 0xc5, 0xd2, 0x94,
	0x82, 0x44, 0xfd, 0xc4, 0xbc, 0xbc, 0xfc, 0x92, 0xc4, 0x92, 0xcc, 0xfc, 0xbc, 0x62, 0xfd, 0xb2,
	0xd4, 0xa2, 0xe2, 0xcc, 0xfc, 0x3c, 0xb8, 0x31, 0x4a, 0x07, 0x18, 0xb9, 0xb8, 0xfd, 0xf2, 0x53,
	0x52, 0x7d, 0x21, 0x46, 0x08, 0xd9, 0x72, 0xb1, 0x83, 0xdc, 0x17, 0x9f, 0x99, 0x22, 0xc1, 0xa8,
	0xc0, 0xa8, 0xc1, 0x6d, 0xa4, 0xa2, 0x87, 0xd5, 0x8d, 0x7a, 0xc1, 0x60, 0xc7, 0x40, 0xb5, 0x05,
	0xb1, 0x81, 0x34, 0x79, 0xa6, 0x08, 0x79, 0x73, 0xf1, 0x81, 0xb5, 0xe7, 0xa6, 0x96, 0x24, 0xa6,
	0x24, 0x96, 0x24, 0x16, 0x4b, 0x30, 0x29, 0x30, 0xe3, 0x37, 0xa5, 0x34, 0xb9, 0x04, 0x66, 0x0a,
	0x2f, 0x48, 0xaf, 0x2f, 0x4c, 0xab, 0x95, 0xea, 0xac, 0xa3, 0x1d, 0x72, 0x0a, 0x5c, 0x72, 0x58,
	0xb4, 0x22, 0x39, 0xd9, 0xc9, 0x88, 0x4b, 0x39, 0x33, 0x1f, 0x62, 0x7e, 0x41, 0x51, 0x7e, 0x45,
	0x25, 0x76, 0xab, 0x9c, 0x38, 0x41, 0x7a, 0x02, 0x40, 0x9e, 0x0e, 0x60, 0x4c, 0x62, 0x03, 0xfb,
	0xde, 0x18, 0x10, 0x00, 0x00, 0xff, 0xff, 0xfa, 0xa4, 0xfe, 0xde, 0xa3, 0x01, 0x00, 0x00,
}
