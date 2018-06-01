package client

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/zuiwuchang/go-intranet-forward/protocol"
	"github.com/zuiwuchang/king-go/net/easy"
)

// Analyze .
type Analyze struct {
}

// Header .
func (Analyze) Header() int {
	return protocol.HeaderSize
}

// Analyze .
func (Analyze) Analyze(b []byte) (int, error) {
	//header
	if len(b) != protocol.HeaderSize {
		return 0, easy.ErrorHeaderSize
	}

	header := protocol.Header(b)

	//flag
	if protocol.HeaderFlag != header.Flag() {
		return 0, easy.ErrorHeaderFlag
	}

	//cmd
	cmd := header.Command()
	if cmd < protocol.CommandBegin && cmd >= protocol.CommandEnd {
		return 0, easy.ErrorHeaderCommand
	}

	//len
	n := header.Len()
	if n < 6 {
		return 0, easy.ErrorMessageSize
	}

	return n, nil
}

// Request .
func Request(c easy.IClient,
	request uint16, requestPB proto.Message,
	reply uint16, replyPB proto.Message,
) (e error) {
	// 發送 請求
	var b protocol.Message
	b, e = protocol.NewMessage(request, requestPB)
	if e != nil {
		return
	}
	_, e = c.Write(b)
	if e != nil {
		return
	}

	// 獲取 回覆
	b, e = c.Read(nil)
	if e != nil {
		return
	}
	// 驗證 回覆
	cmd := b.Command()
	if cmd != reply {
		e = fmt.Errorf("reply expect to get %v instead of %v", reply, cmd)
		return
	}
	if replyPB != nil {
		e = b.Body(replyPB)
	}
	return
}
