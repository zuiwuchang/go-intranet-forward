package protocol

import (
	"encoding/binary"
	"github.com/golang/protobuf/proto"
)

const (
	// HeaderSize .
	HeaderSize = 6

	// HeaderFlag .
	HeaderFlag = 1911
)

const (
	// CommandBegin .
	CommandBegin = 1

	// Register 向服務器 註冊映射
	Register = iota + CommandBegin
	// RegisterOK 註冊映射成功
	RegisterOK

	// CommandEnd .
	CommandEnd
)

// ByteOrder .
var ByteOrder = binary.LittleEndian

// Message .
type Message []byte

// Header .
type Header []byte

// Command .
func (b Header) Command() uint16 {
	return ByteOrder.Uint16(b[2:])
}

// Command .
func (b Message) Command() uint16 {
	return ByteOrder.Uint16(b[2:])
}

// Flag .
func (b Header) Flag() uint16 {
	return ByteOrder.Uint16(b)
}

// Flag .
func (b Message) Flag() uint16 {
	return ByteOrder.Uint16(b)
}

// Len .
func (b Header) Len() int {
	return int(ByteOrder.Uint16(b[4:]))
}

// Len .
func (b Message) Len() int {
	return int(ByteOrder.Uint16(b[4:]))
}

// Body 返回 解碼後的 body 到 pb 中
func (b Message) Body(pb proto.Message) (e error) {
	proto.Unmarshal(b, pb)
	return
}
