package client

import (
	"errors"
	"github.com/zuiwuchang/go-intranet-forward/configure"
	"github.com/zuiwuchang/go-intranet-forward/log"
	"github.com/zuiwuchang/go-intranet-forward/protocol"
	"github.com/zuiwuchang/go-intranet-forward/protocol/go/pb"
	"github.com/zuiwuchang/king-go/net/easy"
	"net"
)

// Logger .
var Logger = log.Logger

// Run .
func Run() {
	srv := configure.GetClient()
	// forward
	keysForward := make(map[uint32]*Forward)
	for i := 0; i < len(srv.Forward); i++ {
		node := srv.Forward[i]
		f, e := NewForward(node)
		if e != nil {
			Logger.Fault.Fatalln(e)
		}
		keysForward[node.ID] = f
	}

	for _, forward := range keysForward {
		client, e := NewClient(forward)
		if e != nil {
			Logger.Fault.Fatalln(e)
		}
		forward.Session, e = NewSession(forward.ID,
			client,
			forward.SendBuffer, forward.Local,
			forward.TunnelRecvBuffer, forward.TunnelSendBuffer,
		)
		if e != nil {
			Logger.Fault.Fatalln(e)
		}
	}

	// 創建服務
	service := Service{
		keysForward: keysForward,
	}
	service.Run()
}

// NewClient .
func NewClient(forward *Forward) (client easy.IClient, e error) {
	var c net.Conn
	c, e = net.Dial("tcp", forward.Remote)
	if e != nil {
		return
	}
	c0 := easy.NewClient(c, forward.RecvBuffer, Analyze{})

	// 發送 請求
	var reply pb.RegisterReply
	e = Request(c0,
		protocol.Register,
		&pb.Register{
			ID:       forward.ID,
			Password: forward.Hash,
		},
		protocol.RegisterReply,
		&reply,
	)
	if e != nil {
		c0.Close()
		return
	} else if reply.Code != 0 {
		e = errors.New(reply.Error)
		c0.Close()
		return
	}

	// 返回 連接
	client = c0
	return
}
