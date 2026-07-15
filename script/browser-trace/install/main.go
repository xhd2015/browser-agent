// Install bundles the Chrome extension into the embed tree, then installs
// browser-trace into $GOBIN or $GOPATH/bin via `go install`.
//
// Usage (from repo root):
//
//	go run ./script/browser-trace/install
//	go run ./script/browser-trace/install --fixture   # mini embed (no npm)
//
// Git tracks only browsertrace/embedded/**/placeholder.txt. Generated payloads
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
	pkgPath = "./cmd/browser-trace"
	binName = "browser-trace"
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

	needBundle := forceBundle || fixture || diskEmbedIncomplete(root)
	if needBundle {
		if !forceBundle && !fixture {
			fmt.Println("==> Embed incomplete (placeholders / missing outstanding files); auto-bundling…")
		} else {
			fmt.Println("==> Bundling Chrome extension into embed tree")
		}
		bundleArgs := []string{"run", "./script/browser-trace/bundle"}
		if fixture {
			bundleArgs = append(bundleArgs, "--fixture")
		}
		if err := cmd.Debug().Dir(root).Run("go", bundleArgs...); err != nil {
			return fmt.Errorf("bundle failed: %w\n  hint: use --fixture or see docs/assets-hydrate.md", err)
		}
	} else {
		fmt.Println("==> On-disk embed already complete; skipping bundle (pass --force-bundle to refresh)")
	}

	fmt.Println("==> Installing browser-trace (go install)")
	if err := cmd.Debug().Dir(root).Run("go", "install", pkgPath); err != nil {
		return fmt.Errorf("go install %s failed: %w", pkgPath, err)
	}

	dest, err := installDest()
	if err != nil {
		fmt.Printf("\nInstalled %s via go install %s\n", binName, pkgPath)
		return nil
	}
	fmt.Printf("\nInstalled %s\n", dest)
	fmt.Printf("Ensure %s is on your PATH.\n", filepath.Dir(dest))
	fmt.Println()
	fmt.Println("Next:")
	fmt.Println("  browser-trace")
	fmt.Println("  browser-trace --install-chrome-extension")
	return nil
}

func diskEmbedIncomplete(root string) bool {
	ext := os.DirFS(filepath.Join(root, "browsertrace", "embedded", "extension"))
	return !browseragent.EmbedCompleteFS(ext, browseragent.AssetKindExtension)
}

func printHelp() {
	fmt.Print(`Usage: go run ./script/browser-trace/install [options]

Stage the Chrome extension into browsertrace/embedded/extension (for go:embed),
then install the browser-trace binary with:

  go install ./cmd/browser-trace

Git tracks only embedded/**/placeholder.txt; generated files are gitignored.
When the on-disk embed is incomplete, install auto-bundles before go install.

The binary lands in $GOBIN if set, otherwise $GOPATH/bin (default ~/go/bin).

Options:
  --fixture, --mini   Stage mini fixture instead of building Chrome-Ext-Capture-API
  --force-bundle      Bundle even if the on-disk embed looks complete
  -h, --help          Show this help
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
			if _, err := os.Stat(filepath.Join(dir, "cmd", "browser-trace")); err != nil {
				return "", fmt.Errorf("%s: cmd/browser-trace not found", dir)
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
