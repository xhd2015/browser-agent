package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/xhd2015/browser-agent/browseragent"
)

func main() {
	// serve needs a cancellable context; HandleCLI's serve path blocks forever
	// with context.Background. For the binary, intercept serve and wire signals.
	args := os.Args[1:]
	if len(args) > 0 && args[0] == "serve" {
		if err := runServe(args[1:]); err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}
		return
	}
	if err := browseragent.HandleCLI(args, nil, os.Stdout, os.Stderr); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

func runServe(args []string) error {
	// Re-use HandleCLI for flag parsing by wrapping with signal context via
	// direct Config build — keep binary thin by calling HandleCLI for non-serve
	// and a small serve path here.
	// Delegate to package API that accepts context would be ideal; for now
	// parse via a one-shot HandleCLI is wrong (blocks forever). Build Config.
	addr := ""
	baseDir := ""
	sessionID := ""
	noOpenChrome := false
	noAgentRun := false
	for i := 0; i < len(args); i++ {
		a := args[i]
		switch {
		case a == "--addr" && i+1 < len(args):
			addr = args[i+1]
			i++
		case len(a) > 7 && a[:7] == "--addr=":
			addr = a[7:]
		case a == "--base-dir" && i+1 < len(args):
			baseDir = args[i+1]
			i++
		case len(a) > 11 && a[:11] == "--base-dir=":
			baseDir = a[11:]
		case a == "--session-id" && i+1 < len(args):
			sessionID = args[i+1]
			i++
		case len(a) > 13 && a[:13] == "--session-id=":
			sessionID = a[13:]
		case a == "--no-open-chrome":
			noOpenChrome = true
		case a == "--no-agent-run":
			noAgentRun = true
		case a == "-h" || a == "--help":
			return browseragent.HandleCLI([]string{"--help"}, nil, os.Stdout, os.Stderr)
		}
	}
	if sessionID == "" {
		sessionID = fmt.Sprintf("sess-%d", os.Getpid())
	}
	if baseDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			home = os.TempDir()
		}
		baseDir = home + "/.tmp/browser-agent"
	}
	cfg := browseragent.Config{
		Addr:         addr,
		BaseDir:      baseDir,
		SessionID:    sessionID,
		NoOpenChrome: noOpenChrome,
		NoAgentRun:   noAgentRun,
		Stdout:       os.Stdout,
		Stderr:       os.Stderr,
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	_, err := browseragent.Run(ctx, cfg)
	return err
}
