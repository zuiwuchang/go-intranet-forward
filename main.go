package main

import (
	"github.com/zuiwuchang/go-intranet-forward/cmd"
	"github.com/zuiwuchang/go-intranet-forward/log"
)

func main() {
	// 設置 版本 信息
	cmd.Version = Version
	// 執行命令
	if e := cmd.Execute(); e != nil {
		log.Logger.Fault.Fatalln(e)
	}
}
