// Install rebuilds extension + React session-page into browseragent embed trees,
// then installs browser-agent into $GOBIN or $GOPATH/bin via `go install`.
//
// Usage (from module root):
//
//	go run ./script/browser-agent/install
//	go run ./script/browser-agent/install --fixture   # mini embed only (tests / offline)
//	go run ./script/browser-agent/install --skip-bundle  # only if a real SPA is already staged
//
// By default install always runs a full bundle (vite session-page + extension)
// so the binary embeds a fresh React app — it does not skip when a mini fixture
// already sits under browseragent/embedded/session-page/.
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
	var skipBundle bool
	var forceBundle bool // kept for compatibility; full mode always bundles unless --skip-bundle
	var extra []string
	for _, a := range args {
		switch a {
		case "-h", "--help":
			printHelp()
			return nil
		case "--fixture", "--mini":
			fixture = true
		case "--skip-bundle":
			skipBundle = true
		case "--force-bundle":
			// Historical flag: full install always rebundles; accept as no-op.
			forceBundle = true
		default:
			extra = append(extra, a)
		}
	}
	if len(extra) > 0 {
		return fmt.Errorf("unrecognized args: %s", strings.Join(extra, " "))
	}
	if fixture && skipBundle {
		return fmt.Errorf("cannot combine --fixture and --skip-bundle")
	}
	_ = forceBundle

	root, err := findModuleRoot()
	if err != nil {
		return err
	}

	sessEmbed := filepath.Join(root, "browseragent", "embedded", "session-page")

	// 1) Stage embed trees.
	switch {
	case fixture:
		fmt.Println("==> Bundling mini fixtures into embed (--fixture)")
		res, berr := browseragent.Bundle(browseragent.BundleOptions{
			Root:       root,
			UseFixture: true,
		})
		if berr != nil {
			return fmt.Errorf("fixture bundle failed: %w", berr)
		}
		fmt.Printf("Staged fixture extension → %s\n", res.ExtensionDir)
		fmt.Printf("Staged fixture session-page → %s\n", res.SessionPageDir)
		fmt.Fprintln(os.Stderr, "warning: --fixture embeds the mini session-page (no React SPA)")

	case skipBundle:
		fmt.Println("==> Skipping bundle (--skip-bundle); verifying on-disk embed is a real SPA")
		if diskEmbedsIncomplete(root) {
			return fmt.Errorf("--skip-bundle: on-disk embed incomplete under browseragent/embedded/; omit --skip-bundle to rebuild")
		}
		if browseragent.SessionPageDirIsMiniFixture(sessEmbed) {
			return fmt.Errorf("--skip-bundle: session-page under %s is still the mini fixture; omit --skip-bundle to run a full vite embed", sessEmbed)
		}
		fmt.Printf("Using existing session-page embed → %s\n", sessEmbed)

	default:
		// Full install: always rebuild so //go:embed picks up a fresh React app.
		// Do not skip when a mini fixture already satisfies weak "complete" checks.
		fmt.Println("==> Full bundle (extension + fresh React session-page) → embed")
		res, berr := browseragent.Bundle(browseragent.BundleOptions{
			Root:       root,
			UseFixture: false,
		})
		if berr != nil {
			return fmt.Errorf("full bundle failed: %w\n  hint: need node + npm/pnpm in react/ for vite; or use --fixture for a mini embed only", berr)
		}
		if res.SessionPageFromFixture || browseragent.SessionPageDirIsMiniFixture(res.SessionPageDir) {
			return fmt.Errorf("refusing to install: session-page is still a mini fixture at %s (React SPA not embedded)", res.SessionPageDir)
		}
		fmt.Printf("Staged extension → %s\n", res.ExtensionDir)
		fmt.Printf("Staged session-page (React) → %s\n", res.SessionPageDir)
		if res.ExtensionFromFixture {
			fmt.Fprintln(os.Stderr, "warning: extension came from fixture; session-page is real SPA")
		}
	}

	// 2) go install into GOBIN / GOPATH/bin (embeds browseragent/embedded/**).
	fmt.Println("==> Installing browser-agent (go install; embeds staged session-page)")
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
	if !fixture {
		fmt.Println()
		fmt.Println("Session page: React SPA embedded (not the mini fixture).")
		fmt.Println("Restart any running daemon so it loads this binary:")
		fmt.Println("  browser-agent serve --stop")
		fmt.Println("  browser-agent session new")
	}
	fmt.Println()
	fmt.Println("Next:")
	fmt.Println("  browser-agent serve")
	fmt.Println("  browser-agent install-chrome-extension")
	fmt.Println("  browser-agent skill --show")
	fmt.Println()
	fmt.Println("Load unpacked extension from the path printed by install-chrome-extension.")
	fmt.Println("Default control port: 43761.")
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

Always rebuilds a full embed (Chrome extension + React session-page via vite)
into browseragent/embedded/**, then:

  go install ./cmd/browser-agent

so the binary //go:embed includes a fresh SPA — not a leftover mini fixture.

Options:
  --fixture, --mini   Stage mini fixtures only (no vite). Not for normal use.
  --skip-bundle       Skip rebuild; require an existing non-fixture SPA on disk.
  --force-bundle      Accepted for compatibility (full mode always bundles).
  -h, --help          Show this help

Default (recommended):
  1. vite build react/ → browseragent/embedded/session-page
  2. stage extension → browseragent/embedded/extension
  3. go install ./cmd/browser-agent

If vite/node is missing, install fails (does not silently embed the fixture SPA).
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
