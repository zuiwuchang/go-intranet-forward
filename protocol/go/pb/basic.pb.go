// Code generated by protoc-gen-go. DO NOT EDIT.
// source: basic.proto

/*
Package pb is a generated protocol buffer package.

It is generated from these files:
	basic.proto

It has these top-level messages:
	Register
	RegisterReply
	Connect
	ConnectReply
	Forward
	TunnelClose
*/
package pb

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

// 註冊成為 穿透 客戶端
type Register struct {
	// 請求服務 編號
	ID uint32 `protobuf:"varint,1,opt,name=ID" json:"ID,omitempty"`
	// 連接密碼
	Password string `protobuf:"bytes,2,opt,name=Password" json:"Password,omitempty"`
}

func (m *Register) Reset()                    { *m = Register{} }
func (m *Register) String() string            { return proto.CompactTextString(m) }
func (*Register) ProtoMessage()               {}
func (*Register) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

func (m *Register) GetID() uint32 {
	if m != nil {
		return m.ID
	}
	return 0
}

func (m *Register) GetPassword() string {
	if m != nil {
		return m.Password
	}
	return ""
}

type RegisterReply struct {
	Code  int32  `protobuf:"varint,1,opt,name=Code" json:"Code,omitempty"`
	Error string `protobuf:"bytes,2,opt,name=Error" json:"Error,omitempty"`
}

func (m *RegisterReply) Reset()                    { *m = RegisterReply{} }
func (m *RegisterReply) String() string            { return proto.CompactTextString(m) }
func (*RegisterReply) ProtoMessage()               {}
func (*RegisterReply) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

func (m *RegisterReply) GetCode() int32 {
	if m != nil {
		return m.Code
	}
	return 0
}

func (m *RegisterReply) GetError() string {
	if m != nil {
		return m.Error
	}
	return ""
}

// 向 穿透 客戶端請求  建立一個 隧道
type Connect struct {
	// 隧道 標識
	ID uint64 `protobuf:"varint,1,opt,name=ID" json:"ID,omitempty"`
}

func (m *Connect) Reset()                    { *m = Connect{} }
func (m *Connect) String() string            { return proto.CompactTextString(m) }
func (*Connect) ProtoMessage()               {}
func (*Connect) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{2} }

func (m *Connect) GetID() uint64 {
	if m != nil {
		return m.ID
	}
	return 0
}

type ConnectReply struct {
	// 隧道 標識
	ID    uint64 `protobuf:"varint,1,opt,name=ID" json:"ID,omitempty"`
	Code  int32  `protobuf:"varint,2,opt,name=Code" json:"Code,omitempty"`
	Error string `protobuf:"bytes,3,opt,name=Error" json:"Error,omitempty"`
}

func (m *ConnectReply) Reset()                    { *m = ConnectReply{} }
func (m *ConnectReply) String() string            { return proto.CompactTextString(m) }
func (*ConnectReply) ProtoMessage()               {}
func (*ConnectReply) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{3} }

func (m *ConnectReply) GetID() uint64 {
	if m != nil {
		return m.ID
	}
	return 0
}

func (m *ConnectReply) GetCode() int32 {
	if m != nil {
		return m.Code
	}
	return 0
}

func (m *ConnectReply) GetError() string {
	if m != nil {
		return m.Error
	}
	return ""
}

// 轉發 隧道數據
type Forward struct {
	// 隧道 標識
	ID   uint64 `protobuf:"varint,1,opt,name=ID" json:"ID,omitempty"`
	Data []byte `protobuf:"bytes,2,opt,name=Data,proto3" json:"Data,omitempty"`
}

func (m *Forward) Reset()                    { *m = Forward{} }
func (m *Forward) String() string            { return proto.CompactTextString(m) }
func (*Forward) ProtoMessage()               {}
func (*Forward) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{4} }

func (m *Forward) GetID() uint64 {
	if m != nil {
		return m.ID
	}
	return 0
}

func (m *Forward) GetData() []byte {
	if m != nil {
		return m.Data
	}
	return nil
}

// 關閉 隧道
type TunnelClose struct {
	// 隧道 標識
	ID uint64 `protobuf:"varint,1,opt,name=ID" json:"ID,omitempty"`
}

func (m *TunnelClose) Reset()                    { *m = TunnelClose{} }
func (m *TunnelClose) String() string            { return proto.CompactTextString(m) }
func (*TunnelClose) ProtoMessage()               {}
func (*TunnelClose) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{5} }

func (m *TunnelClose) GetID() uint64 {
	if m != nil {
		return m.ID
	}
	return 0
}

func init() {
	proto.RegisterType((*Register)(nil), "pb.Register")
	proto.RegisterType((*RegisterReply)(nil), "pb.RegisterReply")
	proto.RegisterType((*Connect)(nil), "pb.Connect")
	proto.RegisterType((*ConnectReply)(nil), "pb.ConnectReply")
	proto.RegisterType((*Forward)(nil), "pb.Forward")
	proto.RegisterType((*TunnelClose)(nil), "pb.TunnelClose")
}

func init() { proto.RegisterFile("basic.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 207 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x6c, 0x90, 0x41, 0x4b, 0x86, 0x40,
	0x10, 0x86, 0x71, 0xfb, 0xbe, 0xb4, 0x51, 0x3b, 0x2c, 0x1d, 0x2c, 0x08, 0x64, 0x4f, 0x5e, 0xea,
	0x12, 0x04, 0x9d, 0xb5, 0xc8, 0x5b, 0x2c, 0xfd, 0x81, 0x55, 0x97, 0x10, 0x64, 0x47, 0x66, 0x37,
	0xa4, 0x7f, 0x1f, 0xad, 0x26, 0x22, 0xde, 0xe6, 0xe5, 0xe5, 0x79, 0x1f, 0x18, 0x88, 0x1b, 0x65,
	0xfb, 0xf6, 0x71, 0x24, 0x74, 0xc8, 0xd9, 0xd8, 0x88, 0x67, 0x88, 0xa4, 0xfe, 0xea, 0xad, 0xd3,
	0xc4, 0xaf, 0x81, 0xd5, 0x55, 0x16, 0xe4, 0x41, 0x91, 0x4a, 0x56, 0x57, 0xfc, 0x0e, 0xa2, 0x0f,
	0x65, 0xed, 0x84, 0xd4, 0x65, 0x2c, 0x0f, 0x8a, 0x2b, 0xb9, 0x66, 0xf1, 0x02, 0xe9, 0x3f, 0x27,
	0xf5, 0x38, 0xfc, 0x70, 0x0e, 0xa7, 0x12, 0x3b, 0xed, 0xf1, 0xb3, 0xf4, 0x37, 0xbf, 0x81, 0xf3,
	0x2b, 0x11, 0xd2, 0x42, 0xcf, 0x41, 0xdc, 0x42, 0x58, 0xa2, 0x31, 0xba, 0x75, 0x1b, 0xe3, 0xe9,
	0xcf, 0x28, 0xde, 0x21, 0x59, 0xaa, 0x79, 0x74, 0xd7, 0xaf, 0x12, 0x76, 0x24, 0xb9, 0xd8, 0x4a,
	0x1e, 0x20, 0x7c, 0x43, 0x9a, 0x14, 0x75, 0x47, 0x23, 0x95, 0x72, 0xca, 0x8f, 0x24, 0xd2, 0xdf,
	0xe2, 0x1e, 0xe2, 0xcf, 0x6f, 0x63, 0xf4, 0x50, 0x0e, 0x68, 0xf5, 0x1e, 0x69, 0x2e, 0xfd, 0xc3,
	0x9e, 0x7e, 0x03, 0x00, 0x00, 0xff, 0xff, 0xc3, 0xbe, 0x37, 0xd8, 0x3f, 0x01, 0x00, 0x00,
}
