package client

import (
	"fmt"
	"github.com/zuiwuchang/go-intranet-forward/log"
	"github.com/zuiwuchang/go-intranet-forward/protocol"
	"github.com/zuiwuchang/go-intranet-forward/protocol/go/pb"
	"net"
	"os"
	"runtime"
)

func (s *Service) runCommand() {
	signal := s.signal
	var cmd string
	var e error
	rs := make(chan interface{})
	for {
		fmt.Print("$>")
		fmt.Scan(&cmd)

		if cmd == "e" {
			if e = signal.Done(CommandClose{}); e != nil {
				Logger.Warn.Println(e)
			}
			break
		} else if cmd == "info" {
			if e = signal.Done(CommandInfo{
				signal: rs,
			}); e != nil {
				Logger.Warn.Println(e)
				continue
			}
			<-rs
		} else if cmd == "service" {
			if e = signal.Done(CommandService{
				signal: rs,
			}); e != nil {
				Logger.Warn.Println(e)
				continue
			}
			<-rs
		} else if cmd == "h" {
			if e = signal.Done(CommandHelp{
				signal: rs,
			}); e != nil {
				Logger.Warn.Println(e)
				continue
			}
			<-rs
		}
	}
}

// CommandRS .
type CommandRS struct {
	signal chan interface{}
}

// CommandClose .
type CommandClose struct{}

// CommandInfo .
type CommandInfo CommandRS

// CommandHelp .
type CommandHelp CommandRS

// CommandService .
type CommandService CommandRS

// CommandSessionRoute 為 Session 路由 read 到的 消息
type CommandSessionRoute struct {
	Session *Session
	Message protocol.Message
}

// CommandSessionDestory Session 已關閉 銷毀她
type CommandSessionDestory struct {
	Session *Session
}

// CommandConnectReplay .
type CommandConnectReplay struct {
	Session *Session
	Conn    net.Conn
	Reply   *pb.ConnectReply
}

// CommandTunnelRoute 為 Tunnel 路由 read 到的 消息
type CommandTunnelRoute struct {
	Session *Session
	Tunnel  *Tunnel
	Message []byte
}

// CommandTunnelDestory Tunnel 已關閉 銷毀她
type CommandTunnelDestory struct {
	Session *Session
	Tunnel  *Tunnel
}

// DoneClose .
func (s *Service) DoneClose(command CommandClose) (e error) {
	os.Exit(0)
	return
}

// DoneInfo .
func (s *Service) DoneInfo(command CommandInfo) (e error) {
	fmt.Println(runtime.GOOS, runtime.GOARCH)
	fmt.Println("NumCPU", runtime.NumCPU())
	fmt.Println("NumCgoCall", runtime.NumCgoCall())
	fmt.Println("NumGoroutine", runtime.NumGoroutine())

	command.signal <- nil
	return
}

// DoneHelp .
func (s *Service) DoneHelp(command CommandHelp) (e error) {
	fmt.Println("e info service h")

	command.signal <- nil
	return
}

// DoneService .
func (s *Service) DoneService(command CommandService) (e error) {
	for _, forward := range s.keysForward {
		fmt.Println(forward)
	}
	command.signal <- nil
	return
}

// DoneSessionRoute .
func (s *Service) DoneSessionRoute(command CommandSessionRoute) (_e error) {
	command.Session.RequestRead(command.Message)
	return
}

// DoneSessionDestory .
func (s *Service) DoneSessionDestory(command CommandSessionDestory) (_e error) {
	// 通知 退出 主控 goroutine
	session := command.Session
	session.Quit()

	// 設置 服務 空閒
	if forward, _ := s.keysForward[session.ID]; forward != nil &&
		forward.Session == session {
		if log.Trace != nil {
			log.Trace.Println("destory forward\n", forward)
		}
		forward.Session = nil
	}
	return
}

// DoneConnectReplay .
func (s *Service) DoneConnectReplay(command CommandConnectReplay) (_e error) {
	if command.Conn == nil {
		// 回覆失敗
		msg, e := protocol.NewMessage(protocol.ConnectReply, command.Reply)
		if e != nil {
			Logger.Fault.Println(e)
			return
		}
		command.Session.RequestWrite(msg)
	} else {
		// 運行 隧道
		if e := command.Session.RequestTunnel(command.Reply.ID, command.Conn); e != nil {
			command.Conn.Close()
			return
		}
	}
	return
}

// DoneTunnelRoute .
func (s *Service) DoneTunnelRoute(command CommandTunnelRoute) (_e error) {
	session := command.Session
	if session.quit {
		return
	}
	tunnel := command.Tunnel
	msg, e := protocol.NewMessage(protocol.Forward, &pb.Forward{
		ID:   tunnel.ID,
		Data: command.Message,
	})
	if e != nil {
		tunnel.Local.Close()
		return
	}
	session.RequestWrite(msg)
	return
}

// DoneTunnelDestory .
func (s *Service) DoneTunnelDestory(command CommandTunnelDestory) (_e error) {
	// 通知 退出 主控 goroutine
	command.Tunnel.Quit()

	// 通知 session 移除 隧道
	command.Session.RequestRemoveTunnel(command.Tunnel)
	return
}
