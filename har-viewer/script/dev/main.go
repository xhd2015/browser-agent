package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/xhd2015/browser-agent/har-viewer/server"
	"github.com/xhd2015/xgo/support/cmd"
)

func main() {
	forward := os.Args[1:]
	if !hasFlag(forward, "--port") {
		forward = append([]string{
			"--port",
			strconv.Itoa(server.DefaultDevBackendPort),
		}, forward...)
	}

	args := append([]string{"run", "./har-viewer/", "--dev"}, forward...)
	err := cmd.Debug().Run("go", args...)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func hasFlag(args []string, name string) bool {
	for _, arg := range args {
		if arg == name || strings.HasPrefix(arg, name+"=") {
			return true
		}
	}
	return false
}