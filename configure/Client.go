package configure

import (
	"encoding/json"
	"github.com/google/go-jsonnet"
	"io/ioutil"
)

// Client 客戶端 配置
type Client struct {
	Forward []*ClientForward
	Log     Log
}

var _Client Client

// GetClient 返回 客戶端配置 單件
func GetClient() *Client {
	return &_Client
}

// InitClient 初始化 客戶端 配置
func InitClient(filename string) (e error) {
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
	e = json.Unmarshal(b, &_Client)
	if e != nil {
		return e
	}

	return
}
