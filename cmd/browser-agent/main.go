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
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	return browseragent.ServeWithContext(ctx, args, nil, os.Stdout, os.Stderr)
}