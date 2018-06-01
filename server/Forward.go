package server

import (
	"fmt"
	"github.com/zuiwuchang/go-intranet-forward/configure"
	"github.com/zuiwuchang/go-intranet-forward/protocol"
	"github.com/zuiwuchang/king-go/net/easy"
)

// Forward 轉發 服務信息
type Forward struct {
	// 服務編號
	ID uint32
	// 公網 地址
	Public string
	// 加密密鑰
	Key string
	// 連接密碼 如果為空 不驗證
	Password string
	Hash     string

	Session  *Session
	Listener easy.IListener
}

// NewForward .
func NewForward(forward *configure.ServerForward) (f *Forward, e error) {
	var hash string
	if forward.Password != "" {
		hash, e = protocol.Hash(forward.Key, forward.Password)
		if e != nil {
			return
		}
	}
	f = &Forward{
		ID:       forward.ID,
		Public:   forward.Public,
		Key:      forward.Key,
		Password: forward.Password,
		Hash:     hash,
	}
	return
}
func (f *Forward) String() (str string) {
	var listener, session interface{}
	if f.Listener != nil {
		listener = f.Listener.Addr()
	}
	if f.Session != nil {
		session = f.Session.Client.RemoteAddr()
	}
	str = fmt.Sprintf(`***	%v	***
Public   = %v
Key      = %v
Password = %v
Hash     = %v
Listener = %v
Session = %v`,
		f.ID, f.Public, f.Key, f.Password,
		f.Hash,
		listener,
		session,
	)
	return
}
