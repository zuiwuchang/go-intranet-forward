package server

import (
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
type commandSessionConnect struct {
	Conn net.Conn
}
type commandSessionRemoveTunnel struct {
	Tunnel *Tunnel
}

// Session 一個 端口映射 客戶端 session
type Session struct {
	// 服務 id
	ID uint32

	Analyze *Analyze
	Client  easy.IClient

	signal     command.ICommanderSignal
	SignalRoot command.ICommanderSignal

	sendQuque *protocol.SendQuque
	quit      bool

	// 所有 內網穿透 隧道
	tunnels  map[uint64]*Tunnel
	tunnelID uint64

	// 隧道 每次 recv 緩存 最大尺寸
	TunnelRecvBuffer int
	// 隧道 每次 send 數據 最大尺寸
	TunnelSendBuffer int
}

// NewSession .
func NewSession(id uint32, signal command.ICommanderSignal,
	analyze *Analyze, client easy.IClient,
	sendBuffer, tunnelRecvBuffer, tunnelSendBuffer int,
) (session *Session, e error) {
	var sendQuque *protocol.SendQuque
	sendQuque, e = protocol.NewSendQuque(client, sendBuffer)
	if e != nil {
		return
	}
	s := &Session{
		ID:         id,
		Client:     client,
		Analyze:    analyze,
		SignalRoot: signal,
		sendQuque:  sendQuque,

		tunnels: make(map[uint64]*Tunnel),

		TunnelRecvBuffer: tunnelRecvBuffer,
		TunnelSendBuffer: tunnelSendBuffer,
	}
	commander := command.New()
	command.RegisterCommander(commander, s, "Done")
	s.signal = command.NewSignal(make(chan interface{}, 10), commander)

	analyze.Register = true
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
func (s *Session) write() {
	e := s.sendQuque.Run()
	if e != nil {
		if log.Warn != nil {
			log.Warn.Println(e)
		}
		s.Client.Close()
	}
}

// Run .
func (s *Session) Run() {
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

// RequestConnect 請求 一個 轉發連接
func (s *Session) RequestConnect(c net.Conn) {
	if s.quit {
		if log.Debug != nil {
			log.Debug.Println("session already quit ignore request connect", s)
		}
		return
	}
	if e := s.signal.Done(commandSessionConnect{
		Conn: c,
	}); e != nil && log.Fault != nil {
		log.Fault.Println(e)
	}
}

// DoneReadMessage 接收到 消息
func (s *Session) DoneReadMessage(cmd commandSessionRead) (_e error) {
	msg := cmd.Message
	code := msg.Command()
	switch code {
	case protocol.ConnectReply:
		var reply pb.ConnectReply
		if e := msg.Body(&reply); e != nil {
			if log.Error != nil {
				log.Error.Println(e)
			}
			s.Client.Close()
			return
		}
		if reply.Code == 0 {
			s.onConnectReplySuccess(&reply)
		} else {
			s.onConnectReplyError(&reply)
		}
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

// DoneWriteMessage 發送 消息
func (s *Session) DoneWriteMessage(cmd commandSessionWrite) (_e error) {
	if e := s.sendQuque.Send(cmd.Message); e != nil && log.Fault != nil {
		log.Fault.Println(e)
	}

	return
}

// DoneConnect .
func (s *Session) DoneConnect(cmd commandSessionConnect) (_e error) {
	// 獲取 唯一 隧道編號
	s.tunnelID++
	for {
		if _, ok := s.tunnels[s.tunnelID]; !ok {
			break
		}
		s.tunnelID++
	}

	// 創建 請求
	msg, e := protocol.NewMessage(protocol.Connect,
		&pb.Connect{
			ID: s.tunnelID,
		},
	)
	if e != nil {
		if log.Fault != nil {
			log.Fault.Println(e)
		}
		cmd.Conn.Close()
		return
	}

	// 發送 請求
	if e := s.sendQuque.Send(protocol.Message(msg)); e != nil {
		if log.Fault != nil {
			log.Fault.Println(e)
		}
		cmd.Conn.Close()
		return
	}

	// 添加 隧道
	s.tunnels[s.tunnelID] = MallocTunnel(s.tunnelID, s, cmd.Conn)
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
func (s *Session) onConnectReplySuccess(reply *pb.ConnectReply) {
	// 查找 隧道
	id := reply.ID
	if tunnel, _ := s.tunnels[id]; tunnel == nil {
		// 隧道已經不存在 通知 客戶端 關閉 隧道
		if e := s.sendTunnelClose(id); e != nil && log.Fault != nil {
			log.Fault.Println(e)
		}
	} else {
		if tunnel.IsInit() {
			if log.Fault != nil {
				log.Fault.Println("repeat create tunnel", tunnel)
			}
			// 通知 客戶端 關閉 隧道
			if e := s.sendTunnelClose(id); e != nil && log.Fault != nil {
				log.Fault.Println(e)
			}
			// 關閉 本地 隧道
			tunnel.Local.Close()
			delete(s.tunnels, id)
		} else {
			// 初始化隧道
			if e := s.initTunnel(id, tunnel); e == nil {
				// 運行 隧道
				go tunnel.Run()
			} else {
				if log.Fault != nil {
					log.Fault.Println(e)
				}

				// 關閉 本地 隧道
				tunnel.Local.Close()
				delete(s.tunnels, id)
			}
		}
	}
}
func (s *Session) initTunnel(id uint64, tunnel *Tunnel) (e error) {
	e = tunnel.Init(s.SignalRoot, s.TunnelRecvBuffer, s.TunnelSendBuffer)
	if e != nil {
		return
	}

	s.tunnels[id] = tunnel
	return
}
func (s *Session) onConnectReplyError(reply *pb.ConnectReply) {
	// 查找 隧道
	id := reply.ID
	if tunnel, _ := s.tunnels[id]; tunnel == nil {
		if log.Error != nil {
			log.Error.Println("not found tunnel", s, id)
		}
	} else {
		// 驗證 狀態
		if tunnel.IsInit() {
			if log.Error != nil {
				log.Error.Println("tunnel already init", s, id)
			}
		} else {
			// 關閉 socket
			tunnel.Local.Close()
			delete(s.tunnels, id)

			if log.Warn != nil {
				log.Warn.Println("tunnel can't create", id, reply.Error)
			}
		}
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
