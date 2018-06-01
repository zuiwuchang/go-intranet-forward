package server

import (
	"fmt"
	"github.com/zuiwuchang/go-intranet-forward/log"
	"github.com/zuiwuchang/go-intranet-forward/protocol"
	"github.com/zuiwuchang/go-intranet-forward/protocol/go/pb"
	"github.com/zuiwuchang/king-go/command"
	"github.com/zuiwuchang/king-go/net/easy"
	"net"
)

type commandSessionRead protocol.Message
type commandSessionWrite protocol.Message

type commandSessionConnect struct {
	Conn net.Conn
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
}

// NewSession .
func NewSession(id uint32, signal command.ICommanderSignal,
	analyze *Analyze, client easy.IClient,
	sendBuffer int,
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
		if tunnel.Status == tunnelWaitConnect {
			tunnel.Local.Close()
		} else {
			Logger.Fault.Println("none")
		}
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
	if e := s.signal.Done(commandSessionRead(msg)); e != nil && log.Fault != nil {
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

	if e := s.signal.Done(commandSessionWrite(msg)); e != nil && log.Fault != nil {
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
func (s *Session) DoneReadMessage(msg commandSessionRead) (_e error) {
	fmt.Println("read", msg)
	return
}

// DoneWriteMessage 接收到 消息
func (s *Session) DoneWriteMessage(msg commandSessionWrite) (_e error) {
	if e := s.sendQuque.Send(protocol.Message(msg)); e != nil && log.Fault != nil {
		log.Fault.Println(e)
	}

	return
}

// DoneConnect .
func (s *Session) DoneConnect(c commandSessionConnect) (_e error) {
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
		c.Conn.Close()
		return
	}

	// 發送 請求
	if e := s.sendQuque.Send(protocol.Message(msg)); e != nil {
		if log.Fault != nil {
			log.Fault.Println(e)
		}
		c.Conn.Close()
		return
	}

	// 添加 隧道
	s.tunnels[s.tunnelID] = &Tunnel{
		Local:  c.Conn,
		Status: tunnelWaitConnect,
	}
	return
}
