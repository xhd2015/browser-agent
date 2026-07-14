// Bundle stages browseragent/embedded/{extension,session-page} for //go:embed.
//
//	go run ./script/browser-agent/bundle [--fixture|--mini]
//
// With --fixture: copies mini fixtures only (no npm / vite).
// Without: builds extension (public→build) and react session-page (vite), then
// stages into embed. Falls back to fixtures only if a side fails.
//
// Prefer: go run ./script/browser-agent/install  (bundle + go install)
package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/xhd2015/browser-agent/browseragent"
)

func main() {
	if err := handle(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func handle(args []string) error {
	useFixture := false
	for _, a := range args {
		switch a {
		case "-h", "--help":
			fmt.Print(`Usage: go run ./script/browser-agent/bundle [options]

Stage Chrome-Ext-Browser-Agent + react session-page into
browseragent/embedded/** for go:embed.

Options:
  --fixture, --mini   Fixtures only (no vite / node)
  -h, --help          Show this help

Without --fixture:
  1. Copy Chrome-Ext-Browser-Agent/public → build → embed
  2. npm/pnpm install + vite build under react/ → embed

Also see: go run ./script/browser-agent/install
`)
			return nil
		case "--fixture", "--mini":
			useFixture = true
		default:
			if strings.HasPrefix(a, "-") {
				return fmt.Errorf("unrecognized flag %s (try --help)", a)
			}
		}
	}

	root, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("resolve working directory: %w", err)
	}

	res, err := browseragent.Bundle(browseragent.BundleOptions{
		Root:       root,
		UseFixture: useFixture,
	})
	if err != nil {
		return err
	}
	if res == nil {
		return fmt.Errorf("Bundle returned nil result")
	}

	fmt.Printf("Staged extension into %s\n", res.ExtensionDir)
	fmt.Printf("Staged session-page into %s\n", res.SessionPageDir)
	if res.UsedFixture {
		fmt.Printf("UsedFixture=true\n")
		fmt.Fprintln(os.Stderr, "hint: full build needs node for react/; extension stages from public/ without npm")
	} else {
		fmt.Printf("UsedFixture=false\n")
	}
	return nil
}
