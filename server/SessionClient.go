package server

import (
	"github.com/zuiwuchang/king-go/net/easy"
)

// SessionClient 一個 端口映射 客戶端 session
type SessionClient struct {
	Analyze *Analyze
	Client  easy.IClient
}
