package server

import (
	"github.com/zuiwuchang/go-intranet-forward/protocol"
	"github.com/zuiwuchang/king-go/net/easy"
)

// Analyze .
type Analyze struct {
	Register bool
}

// Header .
func (a *Analyze) Header() int {
	return protocol.HeaderSize
}

// Analyze .
func (a *Analyze) Analyze(b []byte) (int, error) {
	//header
	if len(b) != protocol.HeaderSize {
		return 0, easy.ErrorHeaderSize
	}

	header := protocol.Header(b)

	//flag
	if protocol.HeaderFlag != header.Flag() {
		return 0, easy.ErrorHeaderFlag
	}

	//cmd
	cmd := header.Command()
	if a.Register {
		if cmd < protocol.CommandBegin && cmd >= protocol.CommandEnd {
			return 0, easy.ErrorHeaderCommand
		}
	} else {
		if cmd != protocol.Register {
			return 0, easy.ErrorHeaderCommand
		}
	}

	//len
	n := header.Len()
	if n < 6 {
		return 0, easy.ErrorMessageSize
	}

	return n, nil
}
