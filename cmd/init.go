package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/zuiwuchang/go-intranet-forward/client"
	"github.com/zuiwuchang/go-intranet-forward/configure"
	"github.com/zuiwuchang/go-intranet-forward/log"
	"github.com/zuiwuchang/go-intranet-forward/server"
	"os"
)

// Version 版本號
var Version string

const (
	// App 程式名
	App = "go-intranet-forward"
)

// Logger .
var Logger = log.Logger

var v bool
var s string
var c string
var t bool
var rootCmd = &cobra.Command{
	Use:   App,
	Short: "golang intranet forward tools",
	Run: func(cmd *cobra.Command, args []string) {
		if v {
			fmt.Println(Version)
		} else if s != "" {
			e := configure.InitServer(s)
			if e == nil {
				// 初始化 日誌
				log.Init(configure.GetServer().Log)

				server.Run(t)
			} else {
				Logger.Fault.Fatalln(e)
			}
		} else if c != "" {
			e := configure.InitClient(c)
			if e == nil {
				// 初始化 日誌
				log.Init(configure.GetClient().Log)

				client.Run(t)
			} else {
				Logger.Fault.Fatalln(e)
			}
		} else {
			fmt.Println(App)
			fmt.Println(Version)
			fmt.Printf(`Use "%v --help" for more information about this program.
`, App)
		}
	},
}

func init() {
	flags := rootCmd.Flags()
	flags.BoolVarP(&v,
		"version",
		"v",
		false,
		"show version",
	)
	flags.BoolVarP(&t,
		"test",
		"t",
		false,
		"test control",
	)
	flags.StringVarP(&s,
		"server", "s",
		"",
		"server configure file,run as server.",
	)
	flags.StringVarP(&c,
		"client", "c",
		"",
		"client configure file,run as client.",
	)
}

// Execute 執行命令
func Execute() error {
	return rootCmd.Execute()
}
func abort() {
	os.Exit(1)
}
