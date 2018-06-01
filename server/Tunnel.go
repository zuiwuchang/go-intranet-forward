package server

import (
	"net"
)

const (
	// 正在 等待 連接
	tunnelWaitConnect = iota
)

// Tunnel 一個 內網穿透 隧道
type Tunnel struct {
	// 本地 連接
	Local  net.Conn
	Remote *Session
	// 狀態
	Status int
}
