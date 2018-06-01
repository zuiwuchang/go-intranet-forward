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
	// RegisterReply 註冊映射成功
	RegisterReply

	// Connect 服務器 向 客戶端 反向 請求建立一個 隧道
	Connect
	// ConnectReply 客戶端 回覆 隧道 建立情況
	ConnectReply

	// Forward 轉發 數據
	Forward

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
	proto.Unmarshal(b[HeaderSize:], pb)
	return
}

// NewMessage .
func NewMessage(commnad uint16, pb proto.Message) (msg Message, e error) {
	// body
	var body []byte
	if pb != nil {
		body, e = proto.Marshal(pb)
		if e != nil {
			return
		}
	}
	msg = make([]byte, HeaderSize+len(body))
	if body != nil {
		copy(msg[HeaderSize:], body)
	}

	// header
	ByteOrder.PutUint16(msg, HeaderFlag)
	ByteOrder.PutUint16(msg[2:], commnad)
	ByteOrder.PutUint16(msg[4:], uint16(len(msg)))
	return
}
