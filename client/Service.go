package client

import (
	"github.com/zuiwuchang/king-go/command"
)

// Service .
type Service struct {
	signal      command.ICommanderSignal
	keysForward map[uint32]*Forward
}

// Run 運行 服務
func (s *Service) Run(t bool) {
	commander := command.New()
	command.RegisterCommander(commander, s, "Done")
	signal := command.NewSignal(make(chan interface{}, 10), commander)
	s.signal = signal
	if t {
		go s.runCommand()
	}
	for _, forward := range s.keysForward {
		go s.initForward(forward)
	}

	var e error
	for {
		e = signal.Run()
		if e == nil {
			// 已經 關閉 退出
			break
		}
	}
}
