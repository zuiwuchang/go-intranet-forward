package client

import (
	"fmt"
	"github.com/zuiwuchang/go-intranet-forward/log"
	"github.com/zuiwuchang/go-intranet-forward/protocol"
	"github.com/zuiwuchang/go-intranet-forward/protocol/go/pb"
	"net"
	"os"
	"runtime"
	"time"
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

// CommandForwardInitError 重建 Forward 失敗
type CommandForwardInitError struct {
	Forward *Forward
}

// CommandForwardInitSuccess 重建 Forward 成功
type CommandForwardInitSuccess struct {
	Forward *Forward
	Session *Session
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
		if log.Info != nil {
			log.Info.Printf("destory forward,after %v reconstruction.\n", forward.waitSleep)
		}
		forward.Session = nil

		// 重連 服務
		forward.waitInit = true
		go s.initForward(forward)
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
func (s *Service) initForward(forward *Forward) {
	time.Sleep(forward.waitSleep)
	if log.Info != nil {
		log.Info.Println("init forward, id =", forward.ID)
	}

	client, e := NewClient(forward)
	if e != nil {
		if log.Fault != nil {
			log.Fault.Println(e)
		}
		// 通知 主控 創建失敗
		s.signal.Done(CommandForwardInitError{
			Forward: forward,
		})
		return
	}

	var session *Session
	session, e = NewSession(forward.ID,
		client,
		forward.SendBuffer, forward.Local,
		forward.TunnelRecvBuffer, forward.TunnelSendBuffer,
	)
	if e != nil {
		if log.Fault != nil {
			log.Fault.Println(e)
		}
		// 通知 主控 創建失敗
		return
	}

	// 通知 主控 建立成功
	s.signal.Done(CommandForwardInitSuccess{
		Forward: forward,
		Session: session,
	})
}

// DoneForwardInitError .
func (s *Service) DoneForwardInitError(cmd CommandForwardInitError) (_e error) {
	forward := cmd.Forward
	// 驗證 是否 過期
	if f, ok := s.keysForward[forward.ID]; !ok || f != forward {
		if log.Debug != nil {
			log.Debug.Println("forward expired\n", forward)
		}
		return
	}

	// 繼續 重建
	wait := forward.waitSleep * 2
	if wait <= forward.maxWaitSleep {
		forward.waitSleep = wait
	}
	if log.Info != nil {
		log.Info.Printf("forward after %v reconstruction.\n", forward.waitSleep)
	}
	go s.initForward(forward)
	return
}

// DoneForwardInitSuccess .
func (s *Service) DoneForwardInitSuccess(cmd CommandForwardInitSuccess) (_e error) {
	forward := cmd.Forward
	session := cmd.Session
	// 驗證 是否 過期
	if f, ok := s.keysForward[forward.ID]; !ok || f != forward {
		if log.Debug != nil {
			log.Debug.Println("forward expired\n", forward)
		}
		session.Client.Close()
		return
	}

	// 設置 成功數據
	s.keysForward[forward.ID] = forward
	forward.waitSleep = time.Second
	forward.waitInit = false
	forward.Session = session
	if log.Info != nil {
		log.Info.Println("forward reconstruction success\n", forward)
	}

	// 運行 session
	go session.Run(s.signal)
	return
}
