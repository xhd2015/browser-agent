package main

import (
	"embed"
	"fmt"
	"os"

	"github.com/xhd2015/browser-agent/har-viewer/run"
	"github.com/xhd2015/browser-agent/har-viewer/server"
)

//go:embed project-api-capture-react/dist
var distFS embed.FS

//go:embed project-api-capture-react/template.html
var templateHTML string

func main() {
	server.Init(distFS, templateHTML)

	err := run.Run(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
