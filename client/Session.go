package client

import (
	"fmt"
	"github.com/zuiwuchang/go-intranet-forward/log"
	"github.com/zuiwuchang/go-intranet-forward/protocol"
	"github.com/zuiwuchang/go-intranet-forward/protocol/go/pb"
	"github.com/zuiwuchang/king-go/command"
	"github.com/zuiwuchang/king-go/net/easy"
	"net"
)

type commandSessionRead struct {
	Message protocol.Message
}
type commandSessionWrite struct {
	Message protocol.Message
}

type commandRunTunnel struct {
	Tunnel *Tunnel
}
type commandSessionRemoveTunnel struct {
	Tunnel *Tunnel
}

// Session .
type Session struct {
	ID     uint32
	Client easy.IClient

	signal     command.ICommanderSignal
	SignalRoot command.ICommanderSignal

	sendQuque *protocol.SendQuque
	quit      bool

	// 本地 連接 地址
	addr string

	// 所有 內網穿透 隧道
	tunnels map[uint64]*Tunnel

	// 隧道 每次 recv 緩存 最大尺寸
	TunnelRecvBuffer int
	// 隧道 每次 send 數據 最大尺寸
	TunnelSendBuffer int
}

// NewSession .
func NewSession(id uint32,
	c easy.IClient, sendBuffer int,
	addr string,
	tunnelRecvBuffer, tunnelSendBuffer int,
) (session *Session, e error) {
	sendQuque, e := protocol.NewSendQuque(c, sendBuffer)
	if e != nil {
		return nil, e
	}
	s := &Session{
		ID:        id,
		Client:    c,
		sendQuque: sendQuque,
		addr:      addr,
		tunnels:   make(map[uint64]*Tunnel),

		TunnelRecvBuffer: tunnelRecvBuffer,
		TunnelSendBuffer: tunnelSendBuffer,
	}

	commander := command.New()
	command.RegisterCommander(commander, s, "Done")
	s.signal = command.NewSignal(make(chan interface{}, 10), commander)

	session = s
	return
}
func (s *Session) read() {
	signal := s.SignalRoot
	c := s.Client
	var e error
	var msg protocol.Message
	for {
		msg, e = c.Read(nil)
		if e != nil {
			c.Close()
			break
		}
		// 通知 主 goroutine 轉發數據
		signal.Done(CommandSessionRoute{
			Session: s,
			Message: msg,
		})
	}
	// 通知 主 goroutine 銷毀 session
	if e = signal.Done(CommandSessionDestory{
		Session: s,
	}); e != nil && log.Fault != nil {
		log.Fault.Println(e)
	}

}

// Run .
func (s *Session) Run(signalRoot command.ICommanderSignal) {
	s.SignalRoot = signalRoot
	// 啟動 讀取 goroutine
	go s.read()
	// 啟動 寫入 goroutine
	go s.write()

	signal := s.signal
	for {
		e := signal.Run()
		if e == nil {
			break
		} else {
			if log.Fault != nil {
				log.Fault.Println(e)
			}
		}
	}

	s.sendQuque.Exit()
}
func (s *Session) write() {
	e := s.sendQuque.Run()
	if e != nil {
		if log.Warn != nil {
			log.Warn.Println(e)
		}
		s.Client.Close()
	}
}

// Quit 通知 Run 退出
func (s *Session) Quit() {
	if s.quit {
		if log.Fault != nil {
			log.Fault.Println("repeat quit")
		}
		return
	}

	// 通知 主控 goroutine 退出
	s.signal.Close()
	s.quit = true

	// 關閉 隧道
	for _, tunnel := range s.tunnels {
		tunnel.Local.Close()
	}
}

// RequestRead 向 session 請求 處理 read 數據
func (s *Session) RequestRead(msg protocol.Message) {
	if s.quit {
		if log.Debug != nil {
			log.Debug.Println("session already quit ignore request read", s)
		}
		return
	}
	if e := s.signal.Done(commandSessionRead{msg}); e != nil && log.Fault != nil {
		log.Fault.Println(e)
	}
}

// RequestWrite 向 session 請求 處理 write 數據
func (s *Session) RequestWrite(msg protocol.Message) {
	if s.quit {
		if log.Debug != nil {
			log.Debug.Println("session already quit ignore request write", s)
		}
		return
	}

	if e := s.signal.Done(commandSessionWrite{msg}); e != nil && log.Fault != nil {
		log.Fault.Println(e)
	}
}

// RequestRemoveTunnel 向 session 請求 移除 隧道
func (s *Session) RequestRemoveTunnel(tunnel *Tunnel) {
	if s.quit {
		if log.Debug != nil {
			log.Debug.Println("session already quit ignore request remove tunnel", s)
		}
		return
	}
	if e := s.signal.Done(commandSessionRemoveTunnel{
		Tunnel: tunnel,
	}); e != nil && log.Fault != nil {
		log.Fault.Println(e)
	}
}

// RequestTunnel 向 session 請求 運行一個 隧道
func (s *Session) RequestTunnel(id uint64, c net.Conn) (e error) {
	if s.quit {
		e = fmt.Errorf("session already quit ignore request write %v", s)
		if log.Debug != nil {
			log.Debug.Println(e)
		}
		return
	}

	tunnel := &Tunnel{
		ID:      id,
		Local:   c,
		Session: s,
	}

	if e = s.signal.Done(commandRunTunnel{
		Tunnel: tunnel,
	}); e != nil && log.Fault != nil {
		log.Fault.Println(e)
	}
	return
}

// DoneReadMessage 接收到 消息
func (s *Session) DoneReadMessage(cmd commandSessionRead) (_e error) {
	msg := cmd.Message
	code := msg.Command()
	switch code {
	case protocol.Connect:
		var request pb.Connect
		if e := msg.Body(&request); e != nil {
			if log.Error != nil {
				log.Error.Println(e)
			}
			s.Client.Close()
			return
		}

		go s.onConnect(&request)
	case protocol.TunnelClose:
		var request pb.Connect
		if e := msg.Body(&request); e != nil {
			if log.Error != nil {
				log.Error.Println(e)
			}
			s.Client.Close()
			return
		}
		if tunnel, _ := s.tunnels[request.ID]; tunnel != nil {
			tunnel.Local.Close()
			delete(s.tunnels, request.ID)
		}
	case protocol.Forward:
		var request pb.Forward
		if e := msg.Body(&request); e != nil {
			if log.Error != nil {
				log.Error.Println(e)
			}
			s.Client.Close()
			return
		}
		s.onForward(&request)
	default:
		if log.Fault != nil {
			log.Fault.Println("unknow commnad", code)
		}
		s.Client.Close()
	}
	return
}

// DoneWriteMessage 接收到 消息
func (s *Session) DoneWriteMessage(cmd commandSessionWrite) (_e error) {
	if e := s.sendQuque.Send(cmd.Message); e != nil && log.Fault != nil {
		log.Fault.Println(e)
	}
	return
}
func (s *Session) onConnect(request *pb.Connect) {
	reply := &pb.ConnectReply{
		ID: request.ID,
	}
	c, e := net.Dial("tcp", s.addr)
	if e != nil {
		reply.Code = -1
		reply.Error = e.Error()
		// 回覆 失敗
		s.SignalRoot.Done(CommandConnectReplay{
			Session: s,
			Reply:   reply,
		})
		return
	}

	// 回覆 成功
	s.SignalRoot.Done(CommandConnectReplay{
		Session: s,
		Conn:    c,
		Reply:   reply,
	})
}

// DoneRunTunnel 運行 隧道
func (s *Session) DoneRunTunnel(cmd commandRunTunnel) (_e error) {
	// 查詢 隧道id是否可用
	tunnel := cmd.Tunnel
	id := tunnel.ID
	if _, ok := s.tunnels[id]; ok {
		emsg := fmt.Sprintf("duplicate tunnel id\n%v\nid=%v", s, id)
		if log.Error != nil {
			log.Error.Println(emsg)
		}
		// 回覆 失敗
		s.SignalRoot.Done(CommandConnectReplay{
			Session: s,
			Reply: &pb.ConnectReply{
				ID:    id,
				Code:  -1,
				Error: emsg,
			},
		})
		// 關閉 本地 socket
		tunnel.Local.Close()
		return
	}
	// 初始化 隧道
	if e := tunnel.Init(s.SignalRoot,
		s.TunnelRecvBuffer, s.TunnelSendBuffer,
	); e != nil {
		if log.Error != nil {
			log.Error.Println(e)
		}
		// 回覆 失敗
		s.SignalRoot.Done(CommandConnectReplay{
			Session: s,
			Reply: &pb.ConnectReply{
				ID:    id,
				Code:  -1,
				Error: e.Error(),
			},
		})
		// 關閉 本地 socket
		tunnel.Local.Close()
		return
	}

	// 運行 隧道
	go tunnel.Run()

	// 回覆 成功
	s.SignalRoot.Done(CommandConnectReplay{
		Session: s,
		Reply: &pb.ConnectReply{
			ID: id,
		},
	})
	s.tunnels[id] = tunnel
	return
}

// DoneRemoveTunnel 移除 隧道
func (s *Session) DoneRemoveTunnel(cmd commandSessionRemoveTunnel) (_e error) {
	t0 := cmd.Tunnel
	t1, _ := s.tunnels[t0.ID]
	if t0 == t1 {
		delete(s.tunnels, t0.ID)
	}
	s.sendTunnelClose(t0.ID)
	return
}
func (s *Session) sendTunnelClose(id uint64) (e error) {
	var msg protocol.Message
	msg, e = protocol.NewMessage(protocol.TunnelClose,
		&pb.TunnelClose{
			ID: id,
		},
	)
	if e != nil {
		return
	}
	e = s.sendQuque.Send(msg)
	return
}
func (s *Session) onForward(request *pb.Forward) {
	// 查找 通道
	id := request.ID
	tunnel, _ := s.tunnels[id]
	if tunnel == nil {
		if log.Warn != nil {
			log.Warn.Println("tunnel not found", id)
		}
	} else {
		if len(request.Data) > 0 {
			// 通知 主控 轉發 write 數據
			s.SignalRoot.Done(CommandTunnelWrite{
				Tunnel: tunnel,
				Data:   request.Data,
			})
		}
	}
}
