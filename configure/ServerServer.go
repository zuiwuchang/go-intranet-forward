package configure

import (
	"time"
)

// ServerServer .
type ServerServer struct {
	// listen 地址
	Addr string
	// 超時斷線 為0 永不超時
	Timeout time.Duration
}

func (s *ServerServer) format() {
	if s.Timeout < 0 {
		s.Timeout = 0
	} else if s.Timeout != 0 {
		s.Timeout *= time.Millisecond
	}
}
