package run

import (
	"fmt"
	"os"
	"strings"

	"github.com/xhd2015/browser-agent/har-viewer/server"

	"github.com/xhd2015/less-gen/flags"
)

const help = `
Usage: har-viewer [options]

Options:
  --dir <path>       Directory to scan for .har files (default: current directory)
  --cli              Print HAR entries to terminal instead of opening browser
  --dev              Use Vite dev server
  --port <port>      Server port (default: auto from 8080)
  --component <name> Serve a single component
  -h, --help         Show this help
`

func Run(args []string) error {
	var devFlag bool
	var cliFlag bool
	var component string
	var harDir string
	var port int
	args, err := flags.
		Bool("--dev", &devFlag).
		Bool("--cli", &cliFlag).
		String("--component", &component).
		String("--dir", &harDir).
		Int("--port", &port).
		Help("-h,--help", help).
		Parse(args)
	if err != nil {
		return err
	}

	if len(args) > 0 {
		return fmt.Errorf("unrecognized extra args: %s", strings.Join(args, " "))
	}

	if component == "list" {
		fmt.Println("Available components: App, HarViewer")
		return nil
	}

	server.SetHARDir(harDir)

	if cliFlag {
		return runCLI()
	}

	if component == "" {
		component = "App"
	}

	var html string
	if !devFlag {
		html, err = server.FormatTemplateHtml(server.FormatOptions{
			Title:     "API Capture Viewer",
			Component: component,
		})
		if err != nil {
			return err
		}
	}
	return server.ServeComponent(port, server.ServeOptions{
		Dev: devFlag,
		Static: server.StaticOptions{
			IndexHtml: html,
		},
		OpenBrowserUrl: func(port int, url string) string {
			if devFlag {
				return fmt.Sprintf("%s/?component=%s", url, component)
			}
			return url
		},
	})
}

func runCLI() error {
	files, err := server.ListHARFiles()
	if err != nil {
		return err
	}
	if len(files) == 0 {
		return fmt.Errorf("no .har files found")
	}

	for _, file := range files {
		if len(files) > 1 {
			fmt.Fprintf(os.Stdout, "=== %s ===\n\n", file)
		}
		entries, err := server.GetEntrySummaries(file)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading %s: %v\n", file, err)
			continue
		}
		server.PrintEntriesTable(os.Stdout, entries)
		if len(files) > 1 {
			fmt.Fprintln(os.Stdout)
		}
	}
	return nil
}
