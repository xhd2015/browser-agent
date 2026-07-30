// Package lib holds shared helpers for script/github/release.
//
// Pre-build stages fat embeds (generate + full bundle + required AMO-signed
// Firefox .xpi matching VERSION.txt) so multi-platform go build ships a usable
// offline binary with permanent Firefox install support.
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
// When skipPreBuild is true, still requires a matching AMO-signed Firefox .xpi
// already available (dist/signed or embed) so releases never ship placeholder-only.
func BuildRelease(skipPreBuild bool) (*release.BuildReleaseResult, error) {
	var pre func() error
	if !skipPreBuild {
		pre = PreBuild
	} else {
		pre = RequireSignedFirefoxXPIOnly
	}
	return release.BuildRelease(BinaryName, pre, DefaultSpecs, release.WithPackagePath(PackagePath))
}

// PreBuild stages VERSION sinks + fat embeds for //go:embed.
// Requires an AMO-signed Firefox .xpi whose package version matches VERSION.txt.
// Does not run firefox/sign (AMO can take hours) — fails with a clear operator hint.
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

	return stageRequiredSignedFirefoxXPI(root)
}

// RequireSignedFirefoxXPIOnly is the --skip-prebuild gate: no generate/bundle,
// but still require a version-matched AMO-signed xpi staged into embed.
func RequireSignedFirefoxXPIOnly() error {
	root, err := findModuleRoot()
	if err != nil {
		return err
	}
	return stageRequiredSignedFirefoxXPI(root)
}

// CheckSignedFirefoxXPI reports whether a release-quality signed xpi exists for
// the current product version (no staging). Used by --dry-run.
func CheckSignedFirefoxXPI() error {
	root, err := findModuleRoot()
	if err != nil {
		return err
	}
	wantVer := productVersion(root)
	_, err = browseragent.FindReleaseFirefoxXPI(root, wantVer)
	return err
}

func stageRequiredSignedFirefoxXPI(root string) error {
	wantVer := productVersion(root)
	fmt.Printf("==> prebuild: require AMO-signed Firefox xpi (version %s)\n", wantVer)
	src, err := browseragent.FindReleaseFirefoxXPI(root, wantVer)
	if err != nil {
		return err
	}
	fmt.Printf("  ok  %s (AMO-signed, version match)\n", src)
	embedDir, err := browseragent.StageFirefoxXPIEmbed(root, src, wantVer)
	if err != nil {
		return fmt.Errorf("stage firefox xpi embed: %w", err)
	}
	fmt.Printf("Staged Firefox signed xpi → %s/browser-agent.xpi\n", embedDir)
	return browseragent.EnsureEmbedPlaceholders(root)
}

func productVersion(root string) string {
	if v := browseragent.ReadProductVersion(root); v != "" {
		return v
	}
	return strings.TrimSpace(browseragent.ClientVersion())
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
