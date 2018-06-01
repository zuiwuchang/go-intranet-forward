package server

import (
	"github.com/zuiwuchang/go-intranet-forward/configure"
	"github.com/zuiwuchang/go-intranet-forward/log"
	"github.com/zuiwuchang/go-intranet-forward/protocol"
	"github.com/zuiwuchang/go-intranet-forward/protocol/go/pb"
	"github.com/zuiwuchang/king-go/command"
	"github.com/zuiwuchang/king-go/net/easy"
	"net"
	"time"
)

// Service .
type Service struct {
	listen      easy.IListener
	keysForward map[uint32]*Forward
	signal      command.ICommanderSignal
}

// Run 運行 服務
func (s *Service) Run() {
	go s.runListen(s.listen)
	for _, forward := range s.keysForward {
		go s.runListenForward(forward.ID, forward.Listener)
	}
	commander := command.New()
	command.RegisterCommander(commander, s, "Done")
	signal := command.NewSignal(make(chan interface{}, 10), commander)
	s.signal = signal
	go s.runCommand()

	var e error
	for {
		e = signal.Run()
		if e == nil {
			// 已經 關閉 退出
			break
		}
	}
}

// 運行 服務器
func (s *Service) runListen(l easy.IListener) {
	var e error
	var c net.Conn
	for {
		c, e = l.Accept()
		if e != nil {
			if log.Warn != nil {
				log.Warn.Println(e)
			}
			if l.Closed() {
				break
			}
		}
		go s.newSessionClient(c)
	}
}

// 運行 轉發 服務
func (s *Service) runListenForward(id uint32, l easy.IListener) {
	var e error
	var c net.Conn
	for {
		c, e = l.Accept()
		if e != nil {
			if log.Warn != nil {
				log.Warn.Println(e)
			}
			if l.Closed() {
				break
			}
			continue
		}
		if e = s.signal.Done(CommandConnect{
			ID:   id,
			Conn: c,
		}); e != nil && log.Fault != nil {
			log.Fault.Println(e)
		}
	}
}

func (s *Service) newSessionClient(c net.Conn) {
	srv := configure.GetServer().Server
	analyze := &Analyze{}
	client := easy.NewClient(c, srv.RecvBuffer, analyze)

	// 讀取 初始消息
	timeout := srv.InitTimeout
	b, e := client.ReadTimeout(timeout, nil)
	if e != nil {
		if log.Warn != nil {
			log.Warn.Println(e)
		}

		client.Close()
		if e == easy.ErrorReadTimeout {
			client.WaitRead()
		}
		return
	}
	msg := protocol.Message(b)
	if msg.Command() == protocol.Register {
		// 驗證 登入
		var request pb.Register
		if e = msg.Body(&request); e == nil {
			// 通知 主服務 建立 映射
			e = s.signal.Done(CommandRegister{
				Request: &request,
				Client:  client,
				Analyze: analyze,
			})
			if e == nil {
				return
			}
			if log.Error != nil {
				log.Error.Println(e)
			}
		} else {
			if log.Error != nil {
				log.Error.Println(e)
			}
		}
	}
	client.Close()
}

func (s *Service) replyError(c easy.IClient, reply *pb.RegisterReply) {
	msg, e := protocol.NewMessage(protocol.RegisterReply, reply)
	if e != nil {
		if log.Fault != nil {
			log.Fault.Println(e)
		}

		c.Close()
		return
	}

	_, e = c.WriteTimeout(msg, time.Second*10)
	c.Close()
	if e == easy.ErrorWriteTimeout {
		c.WaitWrite()
	}
}
