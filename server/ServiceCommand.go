package server

import (
	"fmt"
	"github.com/zuiwuchang/go-intranet-forward/protocol/go/pb"
	"github.com/zuiwuchang/king-go/net/easy"
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
			}
			<-rs
		} else if cmd == "service" {
			if e = signal.Done(CommandService{
				signal: rs,
			}); e != nil {
				Logger.Warn.Println(e)
			}
			<-rs
		} else if cmd == "h" {
			if e = signal.Done(CommandHelp{
				signal: rs,
			}); e != nil {
				Logger.Warn.Println(e)
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

// DoneClose .
func (s *Service) DoneClose(command CommandClose) (e error) {
	s.signal.Close()
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
		forward.Display()
	}
	command.signal <- nil
	return
}
