// Install rebuilds extension + React session-page into browseragent embed trees,
// stages a signed Firefox .xpi when possible, then installs browser-agent into
// $GOBIN or $GOPATH/bin via `go install`.
//
// Usage (from module root):
//
//	go run ./script/browser-agent/install
//	go run ./script/browser-agent/install --fixture   # mini embed only (tests / offline)
//	go run ./script/browser-agent/install --skip-bundle  # only if a real SPA is already staged
//	go run ./script/browser-agent/install --skip-firefox-sign  # do not call AMO sign
//
// Always runs go run ./script/generate first (root VERSION.txt → sinks).
// By default install always runs a full bundle (vite session-page + extension)
// so the binary embeds a fresh React app — it does not skip when a mini fixture
// already sits under browseragent/embedded/session-page/.
//
// Firefox signed .xpi: after bundle, install stages
// browseragent/embedded/firefox-xpi/browser-agent.xpi from dist/signed/ when
// present, or runs script/browser-agent/firefox/sign when AMO secrets are
// available (unless --skip-firefox-sign / --fixture).
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
	var skipFirefoxSign bool
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
		case "--skip-firefox-sign":
			skipFirefoxSign = true
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

	// 0) Stamp root VERSION.txt into package sinks before bundle / go install.
	fmt.Println("==> generate (sync VERSION.txt)")
	if err := cmd.Debug().Dir(root).Run("go", "run", "./script/generate"); err != nil {
		return fmt.Errorf("generate failed: %w", err)
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
		// Ensure //go:embed firefox-xpi root is non-empty.
		if err := browseragent.EnsureFirefoxXPIPlaceholder(root); err != nil {
			return fmt.Errorf("firefox-xpi placeholder: %w", err)
		}

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

	// 1b) Stage signed Firefox .xpi into //go:embed (permanent install path).
	if !fixture {
		if err := stageFirefoxXPIForInstall(root, skipFirefoxSign); err != nil {
			fmt.Fprintf(os.Stderr, "warning: firefox signed xpi not staged: %v\n", err)
			if err := browseragent.EnsureFirefoxXPIPlaceholder(root); err != nil {
				return fmt.Errorf("firefox-xpi placeholder: %w", err)
			}
		}
	}

	// 2) go install into GOBIN / GOPATH/bin (embeds browseragent/embedded/**).
	fmt.Println("==> Installing browser-agent (go install; embeds staged session-page + firefox xpi)")
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
	fmt.Println("  browser-agent install-firefox-extension   # opens signed .xpi in Firefox when TTY")
	fmt.Println("  browser-agent skill --show")
	fmt.Println()
	fmt.Println("Load unpacked extension from the path printed by install-chrome-extension.")
	fmt.Println("Default control port: 43761.")
	return nil
}

// stageFirefoxXPIForInstall stages a signed .xpi into
// browseragent/embedded/firefox-xpi/ for //go:embed.
//
// Order:
//  1. If dist/signed has an .xpi, stage it (fast path after a prior sign).
//  2. Else if !skipSign and secrets look available, run firefox/sign then stage.
//  3. Else if embed already has a real browser-agent.xpi, keep it.
//  4. Else error (caller may fall back to placeholder).
func stageFirefoxXPIForInstall(root string, skipSign bool) error {
	signedDir := filepath.Join(root, "dist", "signed")
	ver := strings.TrimSpace(browseragent.ClientVersion())

	tryStage := func(src string) error {
		embedDir, err := browseragent.StageFirefoxXPIEmbed(root, src, ver)
		if err != nil {
			return err
		}
		fmt.Printf("Staged Firefox signed xpi → %s/browser-agent.xpi (from %s)\n", embedDir, src)
		return nil
	}

	// 1) Existing signed artifact
	if src, err := browseragent.FindSignedXPIUnder(signedDir); err == nil {
		fmt.Println("==> Staging Firefox signed .xpi from dist/signed")
		return tryStage(src)
	}

	// 2) Sign via AMO when allowed
	if !skipSign {
		fmt.Println("==> Signing Firefox extension (AMO unlisted) → embed")
		if err := cmd.Debug().Dir(root).Run("go", "run", "./script/browser-agent/firefox/sign"); err != nil {
			// Fall through to existing embed / error
			fmt.Fprintf(os.Stderr, "warning: firefox sign failed: %v\n", err)
		} else if src, err := browseragent.FindSignedXPIUnder(signedDir); err == nil {
			return tryStage(src)
		}
	} else {
		fmt.Println("==> Skipping Firefox AMO sign (--skip-firefox-sign)")
	}

	// 3) Keep already-staged real xpi under embed
	embedXPI := filepath.Join(root, "browseragent", "embedded", "firefox-xpi", "browser-agent.xpi")
	if st, err := os.Stat(embedXPI); err == nil && !st.IsDir() && st.Size() >= 4 {
		f, oerr := os.Open(embedXPI)
		if oerr == nil {
			var head [2]byte
			_, _ = f.Read(head[:])
			_ = f.Close()
			if head[0] == 'P' && head[1] == 'K' {
				fmt.Printf("Using existing embedded Firefox xpi → %s\n", embedXPI)
				return nil
			}
		}
	}

	return fmt.Errorf("no signed .xpi in dist/signed and embed incomplete (sign with: go run ./script/browser-agent/firefox/sign)")
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
into browseragent/embedded/**, stages a signed Firefox .xpi when possible, then:

  go install ./cmd/browser-agent

so the binary //go:embed includes a fresh SPA + Firefox xpi — not a leftover mini fixture.

Options:
  --fixture, --mini        Stage mini fixtures only (no vite / no AMO sign).
  --skip-bundle            Skip rebuild; require an existing non-fixture SPA on disk.
  --skip-firefox-sign      Do not call AMO web-ext sign; reuse dist/signed or embed.
  --force-bundle           Accepted for compatibility (full mode always bundles).
  -h, --help               Show this help

Default (recommended):
  1. vite build react/ → browseragent/embedded/session-page
  2. stage Chrome + Firefox extensions → browseragent/embedded/
  3. sign/stage Firefox .xpi → browseragent/embedded/firefox-xpi/browser-agent.xpi
  4. go install ./cmd/browser-agent

Firefox sign needs .firefox-secretes.txt (AMO JWT). Without secrets, install
reuses dist/signed/*.xpi or an already-staged embed xpi when present.

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
