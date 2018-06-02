package client

import (
	"fmt"
	"github.com/zuiwuchang/go-intranet-forward/configure"
	"github.com/zuiwuchang/go-intranet-forward/protocol"
)

// Forward 轉發 服務信息
type Forward struct {
	// 服務編號
	ID uint32
	// 遠端地址
	Remote string
	// 本機地址
	Local string
	// 加密密鑰
	Key string
	// 連接密碼
	Password string
	Hash     string

	// 每次 recv 緩存 最大尺寸
	RecvBuffer int
	// 每次 send 數據 最大尺寸
	SendBuffer int
	// 隧道 每次 recv 緩存 最大尺寸
	TunnelRecvBuffer int
	// 隧道 每次 send 數據 最大尺寸
	TunnelSendBuffer int

	Session *Session
}

// NewForward .
func NewForward(forward *configure.ClientForward) (f *Forward, e error) {
	var hash string
	hash, e = protocol.Hash(forward.Key, forward.Password)
	if e != nil {
		return
	}
	f = &Forward{
		ID:               forward.ID,
		Remote:           forward.Remote,
		Local:            forward.Local,
		Key:              forward.Key,
		Password:         forward.Password,
		Hash:             hash,
		RecvBuffer:       forward.RecvBuffer,
		SendBuffer:       forward.SendBuffer,
		TunnelRecvBuffer: forward.TunnelRecvBuffer,
		TunnelSendBuffer: forward.TunnelSendBuffer,
	}

	if f.RecvBuffer < 1024 {
		f.RecvBuffer = 1024 * 16
	}
	if f.SendBuffer < 1024 {
		f.SendBuffer = 1024 * 16
	}
	if f.TunnelRecvBuffer < 1024 {
		f.TunnelRecvBuffer = 1024 * 16
	}
	if f.TunnelSendBuffer < 1024 {
		f.TunnelSendBuffer = 1024 * 16
	}
	return
}

// Display .
func (f *Forward) String() (str string) {
	var session interface{}
	if f.Session != nil {
		session = f.Session.Client.LocalAddr()
	}
	str = fmt.Sprintf(`***	%v	***
Remote   = %v
Local    = %v
Key      = %v
Password = %v
Hash     = %v
Session = %v`,
		f.ID, f.Remote, f.Local, f.Key, f.Password,
		f.Hash,
		session,
	)
	return
}
