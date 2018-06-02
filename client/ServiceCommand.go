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

// CommandTunnelWrite 路由 write 到 Tunnel 的數據
type CommandTunnelWrite struct {
	Tunnel *Tunnel
	Data   []byte
}

// DoneClose .
func (s *Service) DoneClose(cmd CommandClose) (e error) {
	os.Exit(0)
	return
}

// DoneInfo .
func (s *Service) DoneInfo(cmd CommandInfo) (e error) {
	fmt.Println(runtime.GOOS, runtime.GOARCH)
	fmt.Println("NumCPU", runtime.NumCPU())
	fmt.Println("NumCgoCall", runtime.NumCgoCall())
	fmt.Println("NumGoroutine", runtime.NumGoroutine())

	cmd.signal <- nil
	return
}

// DoneHelp .
func (s *Service) DoneHelp(cmd CommandHelp) (e error) {
	fmt.Println("e info service h")

	cmd.signal <- nil
	return
}

// DoneService .
func (s *Service) DoneService(cmd CommandService) (e error) {
	for _, forward := range s.keysForward {
		fmt.Println(forward)
	}
	cmd.signal <- nil
	return
}

// DoneSessionRoute .
func (s *Service) DoneSessionRoute(cmd CommandSessionRoute) (_e error) {
	cmd.Session.RequestRead(cmd.Message)
	return
}

// DoneSessionDestory .
func (s *Service) DoneSessionDestory(cmd CommandSessionDestory) (_e error) {
	// 通知 退出 主控 goroutine
	session := cmd.Session
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
func (s *Service) DoneConnectReplay(cmd CommandConnectReplay) (_e error) {
	if cmd.Conn == nil {
		// 回覆失敗
		msg, e := protocol.NewMessage(protocol.ConnectReply, cmd.Reply)
		if e != nil {
			Logger.Fault.Println(e)
			return
		}
		cmd.Session.RequestWrite(msg)
	} else {
		// 運行 隧道
		if e := cmd.Session.RequestTunnel(cmd.Reply.ID, cmd.Conn); e != nil {
			cmd.Conn.Close()
			return
		}
	}
	return
}

// DoneTunnelRoute .
func (s *Service) DoneTunnelRoute(cmd CommandTunnelRoute) (_e error) {
	session := cmd.Session
	if session.quit {
		if log.Debug != nil {
			log.Debug.Println("session already quit ignore request write", session)
		}
		return
	}
	tunnel := cmd.Tunnel
	msg, e := protocol.NewMessage(protocol.Forward, &pb.Forward{
		ID:   tunnel.ID,
		Data: cmd.Message,
	})
	if e != nil {
		tunnel.Local.Close()
		return
	}
	session.RequestWrite(msg)
	return
}

// DoneTunnelDestory .
func (s *Service) DoneTunnelDestory(cmd CommandTunnelDestory) (_e error) {
	// 通知 退出 主控 goroutine
	cmd.Tunnel.Quit()

	// 通知 session 移除 隧道
	cmd.Session.RequestRemoveTunnel(cmd.Tunnel)
	return
}

// DoneTunnelWrite .
func (s *Service) DoneTunnelWrite(cmd CommandTunnelWrite) (_e error) {
	cmd.Tunnel.RequestWrite(cmd.Data)
	return
}
