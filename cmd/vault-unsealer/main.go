package main

import (
	"os"

	"kubevault.dev/unsealer/pkg/cmds"

	logs "github.com/appscode/go/log/golog"
)

func main() {
	logs.InitLogs()
	defer logs.FlushLogs()

	if err := cmds.NewRootCmd().Execute(); err != nil {
		os.Exit(1)
	}
	os.Exit(0)
}
