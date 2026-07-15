// Install bundles extension + session-page into browseragent embed trees, then
// installs browser-agent into $GOBIN or $GOPATH/bin via `go install`.
//
// Usage (from module root):
//
//	go run ./script/browser-agent/install
//	go run ./script/browser-agent/install --fixture   # mini embed (no npm / vite)
//
// Git tracks only browseragent/embedded/**/placeholder.txt. Generated payloads
// are gitignored; this install auto-bundles when the on-disk embed is incomplete.
package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/xhd2015/browser-agent/browseragent"
	"github.com/xhd2015/xgo/support/cmd"
)

const (
	pkgPath = "./cmd/browser-agent"
	binName = "browser-agent"
)

func main() {
	if err := handle(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func handle(args []string) error {
	var fixture bool
	var forceBundle bool
	var extra []string
	for _, a := range args {
		switch a {
		case "-h", "--help":
			printHelp()
			return nil
		case "--fixture", "--mini":
			fixture = true
		case "--force-bundle":
			forceBundle = true
		default:
			extra = append(extra, a)
		}
	}
	if len(extra) > 0 {
		return fmt.Errorf("unrecognized args: %s", strings.Join(extra, " "))
	}

	root, err := findModuleRoot()
	if err != nil {
		return err
	}

	// 1) Auto-bundle when on-disk embeds are incomplete (placeholders only), or when forced.
	needBundle := forceBundle || fixture || diskEmbedsIncomplete(root)
	if needBundle {
		if !forceBundle && !fixture {
			fmt.Println("==> Embed incomplete (placeholders / missing outstanding files); auto-bundling…")
		} else {
			fmt.Println("==> Bundling browser-agent embed (extension + session-page)")
		}
		bundleArgs := []string{"run", "./script/browser-agent/bundle"}
		if fixture {
			bundleArgs = append(bundleArgs, "--fixture")
		}
		if err := cmd.Debug().Dir(root).Run("go", bundleArgs...); err != nil {
			return fmt.Errorf("bundle failed: %w\n  hint: fix node/vite or use --fixture; or hydrate at runtime (docs/assets-hydrate.md)", err)
		}
	} else {
		fmt.Println("==> On-disk embed already complete; skipping bundle (pass --force-bundle to refresh)")
	}

	// 2) go install into GOBIN / GOPATH/bin.
	fmt.Println("==> Installing browser-agent (go install)")
	if err := cmd.Debug().Dir(root).Run("go", "install", pkgPath); err != nil {
		return fmt.Errorf("go install %s failed: %w", pkgPath, err)
	}

	dest, err := installDest()
	if err != nil {
		fmt.Printf("\nInstalled %s via go install %s\n", binName, pkgPath)
	} else {
		fmt.Printf("\nInstalled %s\n", dest)
		fmt.Printf("Ensure %s is on your PATH.\n", filepath.Dir(dest))
	}
	fmt.Println()
	fmt.Println("Next:")
	fmt.Println("  browser-agent serve")
	fmt.Println("  browser-agent install-chrome-extension")
	fmt.Println("  browser-agent skill --show")
	fmt.Println()
	fmt.Println("Load unpacked extension from the path printed by install-chrome-extension")
	fmt.Println("(or Chrome-Ext-Browser-Agent/build after bundle). Default control port: 43761.")
	return nil
}

// diskEmbedsIncomplete reports whether browseragent/embedded trees lack outstanding files.
func diskEmbedsIncomplete(root string) bool {
	ext := os.DirFS(filepath.Join(root, "browseragent", "embedded", "extension"))
	sess := os.DirFS(filepath.Join(root, "browseragent", "embedded", "session-page"))
	return !browseragent.EmbedCompleteFS(ext, browseragent.AssetKindExtension) ||
		!browseragent.EmbedCompleteFS(sess, browseragent.AssetKindSessionPage)
}

func printHelp() {
	fmt.Print(`Usage: go run ./script/browser-agent/install [options]

Stage Chrome-Ext-Browser-Agent + react session-page into
browseragent/embedded/** (for go:embed), then install the binary:

  go install ./cmd/browser-agent

Git tracks only embedded/**/placeholder.txt; generated files are gitignored.
When the on-disk embed is incomplete (placeholders only), install auto-bundles
before go install.

The binary lands in $GOBIN if set, otherwise $GOPATH/bin (default ~/go/bin).

Options:
  --fixture, --mini   Stage mini fixtures only (no vite / node build)
  --force-bundle      Bundle even if the on-disk embed looks complete
  -h, --help          Show this help

Without --fixture, bundle will:
  1. Copy Chrome-Ext-Browser-Agent/public → build → embed
  2. npm/pnpm install + vite build under react/ → embed
  3. go install ./cmd/browser-agent
`)
}

func findModuleRoot() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("resolve working directory: %w", err)
	}
	dir := wd
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			if _, err := os.Stat(filepath.Join(dir, "cmd", "browser-agent")); err != nil {
				return "", fmt.Errorf("%s: cmd/browser-agent not found (not module root?)", dir)
			}
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("go.mod not found from %s; run from the module root", wd)
		}
		dir = parent
	}
}

func installDest() (string, error) {
	if gobin := strings.TrimSpace(os.Getenv("GOBIN")); gobin != "" {
		return filepath.Join(gobin, binName), nil
	}
	gopath := strings.TrimSpace(os.Getenv("GOPATH"))
	if gopath == "" {
		out, err := exec.Command("go", "env", "GOPATH").Output()
		if err != nil {
			return "", err
		}
		gopath = strings.TrimSpace(string(out))
	}
	if gopath == "" {
		return "", fmt.Errorf("GOPATH empty")
	}
	if i := strings.IndexByte(gopath, filepath.ListSeparator); i >= 0 {
		gopath = gopath[:i]
	}
	return filepath.Join(gopath, "bin", binName), nil
}
