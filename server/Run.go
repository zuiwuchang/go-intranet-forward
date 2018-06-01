package server

import (
	"github.com/zuiwuchang/go-intranet-forward/configure"
	"github.com/zuiwuchang/go-intranet-forward/log"
	"github.com/zuiwuchang/king-go/net/easy"
	"net"
)

// Logger .
var Logger = log.Logger

// Run .
func Run() {
	srv := configure.GetServer()

	// server
	l, e := net.Listen("tcp", srv.Server.Addr)
	if e == nil {
		if log.Info != nil {
			log.Info.Println("work at", srv.Server.Addr)
		}
	} else {
		Logger.Fault.Fatalln(e)
	}

	// forward
	keys := make(map[uint32]*Forward)
	for i := 0; i < len(srv.Forward); i++ {
		node := srv.Forward[i]
		f, e := NewForward(node)
		if e != nil {
			Logger.Fault.Fatalln(e)
		}
		keys[node.ID] = f
	}
	var lf net.Listener
	for _, node := range keys {
		lf, e = net.Listen("tcp", node.Public)
		if e == nil {
			if log.Info != nil {
				log.Info.Println("forward", node.Public)
			}
			node.Listener = easy.NewListener(lf)
		} else {
			Logger.Fault.Fatalln(e)
		}
	}

	service := Service{
		listen:      easy.NewListener(l),
		keysForward: keys,
	}
	service.Run()
}
