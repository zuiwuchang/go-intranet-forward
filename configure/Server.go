package configure

import (
	"encoding/json"
	"github.com/google/go-jsonnet"
	"io/ioutil"
)

// Server 服務器 配置
type Server struct {
	Server  ServerServer
	Forward []*ServerForward
	Log     Log
}

var _Server Server

// GetServer 返回 服務器配置 單件
func GetServer() *Server {
	return &_Server
}

// InitServer 初始化 服務器 配置
func InitServer(filename string) (e error) {
	// jsonnet
	b, e := ioutil.ReadFile(filename)
	if e != nil {
		return e
	}

	vm := jsonnet.MakeVM()
	str, e := vm.EvaluateSnippet("", string(b))
	if e != nil {
		return e
	}
	b = []byte(str)

	// json
	e = json.Unmarshal(b, &_Server)
	if e != nil {
		return e
	}

	_Server.Server.format()
	return
}
