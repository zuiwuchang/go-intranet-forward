package server

import (
	"github.com/zuiwuchang/go-intranet-forward/log"
	"github.com/zuiwuchang/go-intranet-forward/protocol"
	"github.com/zuiwuchang/king-go/command"
	"net"
)

// Tunnel 一個 內網穿透 隧道
type Tunnel struct {
	ID uint64

	// 本地 連接
	Local   net.Conn
	Session *Session

	signal     command.ICommanderSignal
	SignalRoot command.ICommanderSignal

	sendQuque *protocol.SendQuque
	quit      bool

	RecvBuffer int
}

// MallocTunnel 申請一個 Tunnel
func MallocTunnel(id uint64, s *Session, c net.Conn) *Tunnel {
	return &Tunnel{
		ID:      id,
		Local:   c,
		Session: s,
	}
}

// IsInit 返回 是否初始化
func (t *Tunnel) IsInit() bool {
	return t.signal != nil
}

// Init .
func (t *Tunnel) Init(signalRoot command.ICommanderSignal, recvBuffer, sendBuffer int) (e error) {
	t.sendQuque, e = protocol.NewSendQuque(t.Local, sendBuffer)
	if e != nil {
		return e
	}

	commander := command.New()
	command.RegisterCommander(commander, t, "Done")
	t.signal = command.NewSignal(make(chan interface{}, 10), commander)

	t.RecvBuffer = recvBuffer

	t.SignalRoot = signalRoot
	return
}

// Run .
func (t *Tunnel) Run() {
	// 啟動 讀取 goroutine
	go t.read()
	// 啟動 寫入 goroutine
	go t.write()

	signal := t.signal
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

	t.sendQuque.Exit()
	return
}
func (t *Tunnel) write() {
	e := t.sendQuque.Run()
	if e != nil {
		if log.Warn != nil {
			log.Warn.Println(e)
		}
		t.Local.Close()
	}
}
func (t *Tunnel) read() {
	signal := t.SignalRoot
	c := t.Local
	var e error
	b := make([]byte, t.RecvBuffer)
	var n int
	for {
		n, e = c.Read(b)
		if e != nil {
			c.Close()
			break
		}
		// 通知 主 goroutine 轉發數據
		msg := make([]byte, n)
		copy(msg, b)
		signal.Done(CommandTunnelRoute{
			Session: t.Session,
			Tunnel:  t,
			Message: msg,
		})
	}
	// 通知 主 goroutine 銷毀 session
	if e = signal.Done(CommandTunnelDestory{
		Session: t.Session,
		Tunnel:  t,
	}); e != nil && log.Fault != nil {
		log.Fault.Println(e)
	}
}

// Quit 通知 Run 退出
func (t *Tunnel) Quit() {
	if t.quit {
		if log.Fault != nil {
			log.Fault.Println("tunnel repeat quit")
		}
		return
	}

	// 通知 主控 goroutine 退出
	t.signal.Close()
	t.quit = true
}
