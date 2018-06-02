package server

import (
	"fmt"
	"github.com/zuiwuchang/go-intranet-forward/configure"
	"github.com/zuiwuchang/go-intranet-forward/log"
	"github.com/zuiwuchang/go-intranet-forward/protocol"
	"github.com/zuiwuchang/go-intranet-forward/protocol/go/pb"
	"github.com/zuiwuchang/king-go/net/easy"
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

// CommandRegister 請求 建立 映射
type CommandRegister struct {
	Request *pb.Register
	Client  easy.IClient
	Analyze *Analyze
}

// CommandSessionRoute 為 Session 路由 read 到的 消息
type CommandSessionRoute struct {
	Session *Session
	Message protocol.Message
}

// CommandSessionDestory Session 已關閉 銷毀她
type CommandSessionDestory struct {
	Session *Session
}

// CommandConnect 請求一個 連接
type CommandConnect struct {
	// 服務 編號
	ID   uint32
	Conn net.Conn
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

// DoneRegister .
func (s *Service) DoneRegister(cmd CommandRegister) (_e error) {
	var reply pb.RegisterReply
	// 查找 服務
	id := cmd.Request.ID
	forward := s.keysForward[id]
	if forward == nil {
		reply.Code = 1
		reply.Error = fmt.Sprintf("forward id not found %v", id)
		if log.Warn != nil {
			log.Warn.Println(reply.Error)
		}
		go s.replyError(cmd.Client, &reply)
		return
	}

	// 驗證 密碼
	if forward.Hash != "" && forward.Hash != cmd.Request.Password {
		reply.Code = 2
		reply.Error = "forward password not match"
		if log.Warn != nil {
			log.Warn.Println(reply.Error)
		}
		go s.replyError(cmd.Client, &reply)
		return
	}
	// 創建 成功 消息
	msg, e := protocol.NewMessage(protocol.RegisterReply, &reply)
	if e != nil {
		if log.Fault != nil {
			log.Fault.Println(e)
		}
		cmd.Client.Close()
		return
	}

	// 將 已經存在的 session 踢下線
	if forward.Session != nil {
		forward.Session.Client.Close()
	}
	// 創建 session
	var session *Session
	session, e = NewSession(id, s.signal,
		cmd.Analyze, cmd.Client, configure.GetServer().Server.SendBuffer,
		forward.TunnelRecvBuffer, forward.TunnelSendBuffer,
	)
	if e != nil {
		if log.Fault != nil {
			log.Fault.Println(e)
		}
		cmd.Client.Close()
		return
	}
	forward.Session = session

	// 運行 session
	go session.Run()
	// 回覆 成功
	session.RequestWrite(msg)

	if log.Trace != nil {
		log.Trace.Println("new forward\n", forward)
	}
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

// DoneCommandConnect .
func (s *Service) DoneCommandConnect(cmd CommandConnect) (_e error) {
	id := cmd.ID
	c := cmd.Conn
	// 查找 服務
	forward, _ := s.keysForward[id]
	if forward == nil {
		if log.Error != nil {
			log.Error.Println("forward not found", id)
		}
		c.Close()
		return
	}
	// 查找 session
	session := forward.Session
	if session == nil {
		if log.Warn != nil {
			log.Warn.Println("forward session not work\n", forward)
		}
		c.Close()
		return
	}

	session.RequestConnect(c)
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
