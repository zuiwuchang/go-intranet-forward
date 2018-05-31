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
	// 連接密碼
	Password string
	Hash     string

	Session  *SessionClient
	Listener easy.IListener
}

// NewForward .
func NewForward(forward *configure.ServerForward) (f *Forward, e error) {
	var hash string
	hash, e = protocol.Hash(forward.Key, forward.Password)
	if e != nil {
		return
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

// Display .
func (f *Forward) Display() {
	fmt.Printf(`***	%v	***
Public   = %v
Key      = %v
Password = %v
Hash     = %v`,
		f.ID, f.Public, f.Key, f.Password,
		f.Hash,
	)
	if f.Listener == nil {
		fmt.Print("\nListener = nil")
	} else {
		fmt.Printf("\nListener = %v", f.Listener.Addr())
	}
	if f.Session == nil {
		fmt.Print("\nSession  = nil")
	} else {
		fmt.Printf("\nSession = %v", f.Session.Client.RemoteAddr())
	}
	fmt.Println()
}
