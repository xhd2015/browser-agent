// usage: go run ./script/github/release [--dry-run] [--skip-prebuild] [--skip-hydrate]
//
// Release fat browser-agent binaries (and hydrate archives) to GitHub Releases.
// Pattern matches github.com/xhd2015/doctest/script/github/release.
//
// Live requirements:
//   - clean git worktree
//   - HEAD is a v* tag (git describe --tags HEAD)
//   - .upload-credentials.json {token, owner, repo}
//   - node/npm for full SPA bundle (unless --skip-prebuild with fat embeds)
//   - AMO-signed Firefox .xpi matching VERSION.txt under dist/signed or embed
//     (run go run ./script/browser-agent/firefox/sign first; AMO may take hours)
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/xhd2015/browser-agent/browseragent"
	githublib "github.com/xhd2015/browser-agent/script/github/lib"
	"github.com/xhd2015/kool/pkgs/github"
	"github.com/xhd2015/kool/pkgs/release"
	"github.com/xhd2015/less-flags"
	"github.com/xhd2015/xgo/support/cmd"
)

const help = `
Usage: go run ./script/github/release [options]

Release browser-agent fat multi-platform binaries to GitHub Releases, and by
default also pack/upload hydrate archives (session-page + extensions).

Asset names (binaries):
  browser-agent-{tag}-{os}-{arch}
  e.g. browser-agent-v0.3.1-darwin-arm64

Firefox (required):
  Prebuild requires an AMO-signed .xpi whose package version matches VERSION.txt
  (under dist/signed/ or browseragent/embedded/firefox-xpi/). Sign offline first —
  Mozilla can take a long time (hours):

    go run ./script/browser-agent/firefox/sign

  Release never auto-runs sign. --skip-prebuild still requires a matching signed xpi.

Options:
  --dry-run         print plan without building or uploading (still checks signed xpi)
  --skip-prebuild   skip generate/bundle; still require matching AMO-signed xpi
  --skip-hydrate    do not pack/upload session-page/extension tar.gz archives
  -h, --help        show this help

Credentials (live only):
  .upload-credentials.json
    {"token":"ghp_...","owner":"xhd2015","repo":"browser-agent"}

Install published binaries:
  curl -fsSL https://raw.githubusercontent.com/xhd2015/browser-agent/master/install.sh | bash
`

func main() {
	if err := handle(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func handle() error {
	var dryRun bool
	var skipPrebuild bool
	var skipHydrate bool
	args, err := lessflags.
		Bool("--dry-run", &dryRun).
		Bool("--skip-prebuild", &skipPrebuild).
		Bool("--skip-hydrate", &skipHydrate).
		Help("-h,--help", help).
		Parse(os.Args[1:])
	if err != nil {
		return err
	}
	if len(args) > 0 {
		return fmt.Errorf("unrecognized extra args: %s", strings.Join(args, " "))
	}

	tag, err := release.GetTag()
	if err != nil {
		if !dryRun {
			return err
		}
		fmt.Fprintf(os.Stderr, "[dry-run] warning: %v\n", err)
		// Prefer VERSION.txt for planned names when not on a tag.
		if v := strings.TrimSpace(browseragent.ClientVersion()); v != "" {
			if !strings.HasPrefix(v, "v") {
				v = "v" + v
			}
			tag = v
		} else {
			tag = "(unknown)"
		}
	}

	creds, err := release.LoadCredentials(".upload-credentials.json")
	if err != nil {
		if !dryRun {
			return err
		}
		fmt.Fprintf(os.Stderr, "[dry-run] warning: %v\n", err)
		creds = &release.Credentials{Owner: "xhd2015", Repo: "browser-agent"}
	}

	plannedBins := githublib.PlannedBinaryNames(tag)
	var plannedHydrate []string
	if !skipHydrate {
		plannedHydrate = browseragent.AssetReleaseNames(tag)
	}

	// Always require matching AMO-signed Firefox xpi (including dry-run).
	if err := githublib.CheckSignedFirefoxXPI(); err != nil {
		return err
	}
	fmt.Println("ok: AMO-signed Firefox xpi matches VERSION.txt")

	if dryRun {
		fmt.Printf("[dry-run] tag: %s\n", tag)
		if skipPrebuild {
			fmt.Println("[dry-run] would skip generate/bundle (--skip-prebuild); still use existing signed xpi")
		} else {
			fmt.Println("[dry-run] would prebuild: generate + full bundle + stage matching signed Firefox xpi")
		}
		for _, name := range plannedBins {
			fmt.Printf("[dry-run] would build: %s\n", name)
		}
		if skipHydrate {
			fmt.Println("[dry-run] would skip hydrate archives (--skip-hydrate)")
		} else {
			for _, name := range plannedHydrate {
				fmt.Printf("[dry-run] would pack hydrate: %s\n", name)
			}
		}
		fmt.Printf("[dry-run] would upload to %s/%s release (creates if 404)\n", creds.Owner, creds.Repo)
		return nil
	}

	// Live: multi-platform fat binaries.
	result, err := githublib.BuildRelease(skipPrebuild)
	if err != nil {
		return err
	}

	client := github.NewReleaseClient(creds.Token, creds.Owner, creds.Repo)
	rel, err := client.GetOrCreateRelease(result.Tag)
	if err != nil {
		return fmt.Errorf("failed to get or create release for tag %s: %v", result.Tag, err)
	}

	for _, file := range result.Files {
		if err := client.UploadReleaseAsset(rel.ID, file); err != nil {
			return fmt.Errorf("failed to upload %s: %v", file, err)
		}
		fmt.Printf("Uploaded %s\n", file)
	}

	if !skipHydrate {
		archives, err := packHydrateArchives(result.Tag)
		if err != nil {
			return fmt.Errorf("pack hydrate: %w", err)
		}
		for _, file := range archives {
			if err := client.UploadReleaseAsset(rel.ID, file); err != nil {
				return fmt.Errorf("failed to upload %s: %v", file, err)
			}
			fmt.Printf("Uploaded %s\n", file)
		}
	}

	fmt.Printf("Release %s: https://github.com/%s/%s/releases/tag/%s\n",
		result.Tag, creds.Owner, creds.Repo, result.Tag)
	return nil
}

// packHydrateArchives runs release-assets pack into a temp dir and returns paths.
func packHydrateArchives(tag string) ([]string, error) {
	outDir, err := os.MkdirTemp("", "browser-agent-release-hydrate-*")
	if err != nil {
		return nil, err
	}
	// Prefer go run so pack logic stays in one place.
	if err := cmd.Debug().Run("go", "run", "./script/github/release-assets",
		"--out", outDir, "--version", tag); err != nil {
		return nil, err
	}
	names := browseragent.AssetReleaseNames(tag)
	var paths []string
	for _, n := range names {
		p := filepath.Join(outDir, n)
		st, err := os.Stat(p)
		if err != nil || st.IsDir() || st.Size() <= 0 {
			return nil, fmt.Errorf("hydrate archive missing after pack: %s", p)
		}
		paths = append(paths, p)
	}
	return paths, nil
}
