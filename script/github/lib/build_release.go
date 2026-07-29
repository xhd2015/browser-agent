// Package lib holds shared helpers for script/github/release.
//
// Pre-build stages fat embeds (generate + full bundle + best-effort signed
// Firefox .xpi) so multi-platform go build ships a usable offline binary.
package lib

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/xhd2015/browser-agent/browseragent"
	"github.com/xhd2015/kool/pkgs/release"
	"github.com/xhd2015/xgo/support/cmd"
)

// DefaultSpecs is the multi-platform matrix (darwin/linux × amd64/arm64).
var DefaultSpecs = release.DefaultSpecs

// BinaryName is the released CLI basename (matches install.sh download formula).
const BinaryName = "browser-agent"

// PackagePath is the go build package for the CLI.
const PackagePath = "./cmd/browser-agent"

// BuildRelease runs pre-build (unless skipPreBuild) then multi-platform builds.
func BuildRelease(skipPreBuild bool) (*release.BuildReleaseResult, error) {
	var pre func() error
	if !skipPreBuild {
		pre = PreBuild
	}
	return release.BuildRelease(BinaryName, pre, DefaultSpecs, release.WithPackagePath(PackagePath))
}

// PreBuild stages VERSION sinks + fat embeds for //go:embed.
// Best-effort Firefox signed xpi: reuses dist/signed or existing embed.
func PreBuild() error {
	root, err := findModuleRoot()
	if err != nil {
		return err
	}
	fmt.Println("==> prebuild: generate (VERSION.txt sinks)")
	if err := cmd.Debug().Dir(root).Run("go", "run", "./script/generate"); err != nil {
		return fmt.Errorf("generate: %w", err)
	}

	fmt.Println("==> prebuild: full bundle (extension + React session-page + firefox embed)")
	res, err := browseragent.Bundle(browseragent.BundleOptions{
		Root:       root,
		UseFixture: false,
	})
	if err != nil {
		return fmt.Errorf("bundle: %w\n  hint: need node + npm/pnpm in react/; or re-run with --skip-prebuild if embeds are already fat", err)
	}
	if res.SessionPageFromFixture || browseragent.SessionPageDirIsMiniFixture(res.SessionPageDir) {
		return fmt.Errorf("refusing release prebuild: session-page is still a mini fixture at %s", res.SessionPageDir)
	}
	fmt.Printf("Staged extension → %s\n", res.ExtensionDir)
	fmt.Printf("Staged session-page → %s\n", res.SessionPageDir)

	// Best-effort signed Firefox .xpi into embed (do not fail release if missing).
	if err := stageFirefoxXPIBestEffort(root); err != nil {
		fmt.Fprintf(os.Stderr, "warning: firefox signed xpi not staged: %v\n", err)
		_ = browseragent.EnsureFirefoxXPIPlaceholder(root)
	}
	return nil
}

func stageFirefoxXPIBestEffort(root string) error {
	signedDir := filepath.Join(root, "dist", "signed")
	ver := strings.TrimSpace(browseragent.ClientVersion())
	if src, err := browseragent.FindSignedXPIUnder(signedDir); err == nil {
		_, err := browseragent.StageFirefoxXPIEmbed(root, src, ver)
		return err
	}
	// Existing real embed xpi is enough for fat binary.
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
	return fmt.Errorf("no signed .xpi in dist/signed and embed incomplete (optional: go run ./script/browser-agent/firefox/sign)")
}

func findModuleRoot() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	dir := wd
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			if _, err := os.Stat(filepath.Join(dir, "cmd", "browser-agent")); err != nil {
				return "", fmt.Errorf("%s: cmd/browser-agent not found", dir)
			}
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("go.mod not found from %s", wd)
		}
		dir = parent
	}
}

// PlannedBinaryNames returns artifact basenames for tag + DefaultSpecs
// (same formula as release.BuildRelease).
func PlannedBinaryNames(tag string) []string {
	var out []string
	for _, spec := range DefaultSpecs {
		out = append(out, fmt.Sprintf("%s-%s-%s-%s", BinaryName, tag, spec.OS, spec.Arch))
	}
	return out
}
