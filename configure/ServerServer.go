package configure

import (
	"time"
)

// ServerServer .
type ServerServer struct {
	// listen 地址
	Addr string
	// 初始化超時時間
	InitTimeout time.Duration

	// 每次 recv 緩存 最大尺寸
	RecvBuffer int
	// 每次 send 數據 最大尺寸
	SendBuffer int
}

func (s *ServerServer) format() {
	if s.InitTimeout < 1 {
		s.InitTimeout = time.Second
	} else {
		s.InitTimeout *= time.Second
	}

	if s.RecvBuffer < 1024 {
		s.RecvBuffer = 1024 * 16
	}
	if s.SendBuffer < 1024 {
		s.SendBuffer = 1024 * 16
	}
}
