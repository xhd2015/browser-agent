// Pre-push helper: refuse push when HEAD commit version stamps disagree.
//
//	go run ./script/git/pre-push
//
// Install (managed hook via git-hooks):
//
//	git-hooks pre-push add 'script.git.pre-push' go run ./script/git/pre-push
//
// Reads blobs from HEAD (git cat-file -p), not the worktree:
//   - VERSION.txt  (product SSoT)
//   - Chrome-Ext-Browser-Agent/build/background.js  (EXT_VERSION fallback string)
//   - Firefox-Ext-Browser-Agent/build/background.js
//   - browseragent/embedded/extension-firefox/background.js
//
// If any path is missing on HEAD, or versions are not identical, prints Error
// on stderr and exits non-zero so the push is aborted.
package main

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/xhd2015/browser-agent/script/internal/clicolor"
)

// Paths relative to repo root, resolved via HEAD:path (not worktree).
const versionFileRel = "VERSION.txt"

var backgroundJSRels = []string{
	"Chrome-Ext-Browser-Agent/build/background.js",
	"Firefox-Ext-Browser-Agent/build/background.js",
	"browseragent/embedded/extension-firefox/background.js",
}

// Matches the ternary fallback after BROWSER_AGENT_BUNDLE_VERSION, e.g.:
//
//	? BROWSER_AGENT_BUNDLE_VERSION
//	: "1.0.2";
//
// Same shape stamped by script/generate into public/background.js (and copied to build/embed).
var reBackgroundVersionFallback = regexp.MustCompile(
	`BROWSER_AGENT_BUNDLE_VERSION[\s\S]{0,200}?:\s*["']([^"']+)["']\s*;`,
)

func main() {
	mode, remain, err := clicolor.ParseFlags(os.Args[1:])
	c := clicolor.NewStyle(mode)
	if err != nil {
		c.ErrorLine(os.Stderr, err.Error())
		os.Exit(1)
	}
	if err := run(remain, c); err != nil {
		c.ErrorLine(os.Stderr, err.Error())
		os.Exit(1)
	}
}

func run(args []string, c clicolor.Style) error {
	for _, a := range args {
		switch a {
		case "-h", "--help":
			fmt.Print(helpText())
			return nil
		default:
			if strings.HasPrefix(a, "-") {
				return fmt.Errorf("unrecognized flag %s (try --help)", a)
			}
			return fmt.Errorf("unexpected argument %q (try --help)", a)
		}
	}

	root, err := gitRoot()
	if err != nil {
		return err
	}

	want, err := readHEADVersionTXT(root)
	if err != nil {
		return err
	}
	fmt.Printf("%s %s %s\n",
		c.Gray("pre-push:"),
		c.Gray("HEAD VERSION.txt ="),
		c.Green(want),
	)

	var mismatches []string
	for _, rel := range backgroundJSRels {
		body, err := gitCatFileHEAD(root, rel)
		if err != nil {
			return fmt.Errorf("%s: %w\n  hint: ensure build/ and extension-firefox embeds are committed after generate + bundle", rel, err)
		}
		got, err := extractBackgroundFallbackVersion(body)
		if err != nil {
			return fmt.Errorf("%s: %w", rel, err)
		}
		token := got
		if got == want {
			token = c.Green(got)
		} else {
			token = c.Red(got)
		}
		fmt.Printf("  %s %s %s\n", c.Gray(rel), c.Gray("→"), token)
		if got != want {
			mismatches = append(mismatches, fmt.Sprintf("%s: got %q, want %q (VERSION.txt)", rel, got, want))
		}
	}

	if len(mismatches) > 0 {
		return fmt.Errorf("HEAD version mismatch — aborting push:\n  %s\n  fix: go run ./script/generate && rebuild extension build/ + firefox embed, commit, then push",
			strings.Join(mismatches, "\n  "))
	}
	fmt.Printf("%s %s\n", c.Gray("pre-push:"), c.Green("versions match"))
	return nil
}

func helpText() string {
	return `Usage: go run ./script/git/pre-push [options]

Verify that the HEAD commit has a consistent product version:
  VERSION.txt
  Chrome-Ext-Browser-Agent/build/background.js   (EXT_VERSION fallback)
  Firefox-Ext-Browser-Agent/build/background.js
  browseragent/embedded/extension-firefox/background.js

Content is read from git (git cat-file -p HEAD:<path>), not the working tree.

Options:
` + clicolor.FlagHelp + `  -h, --help              Show this help

Exit 0 when all versions match; non-zero aborts the push when used as a pre-push hook.
`
}

func readHEADVersionTXT(root string) (string, error) {
	raw, err := gitCatFileHEAD(root, versionFileRel)
	if err != nil {
		return "", fmt.Errorf("VERSION.txt: %w", err)
	}
	ver := strings.TrimSpace(string(raw))
	ver = strings.TrimPrefix(ver, "v")
	ver = strings.TrimPrefix(ver, "V")
	ver = strings.TrimSpace(ver)
	if ver == "" {
		return "", fmt.Errorf("VERSION.txt on HEAD is empty")
	}
	if strings.ContainsAny(ver, "/\\ \t\n\r") {
		return "", fmt.Errorf("VERSION.txt on HEAD has invalid version %q", ver)
	}
	return ver, nil
}

// extractBackgroundFallbackVersion returns the string after BROWSER_AGENT_BUNDLE_VERSION ?: "…".
func extractBackgroundFallbackVersion(src []byte) (string, error) {
	m := reBackgroundVersionFallback.FindSubmatch(src)
	if m == nil {
		return "", fmt.Errorf("no BROWSER_AGENT_BUNDLE_VERSION fallback version string found")
	}
	v := strings.TrimSpace(string(m[1]))
	if v == "" {
		return "", fmt.Errorf("empty fallback version string")
	}
	return v, nil
}

func gitCatFileHEAD(root, rel string) ([]byte, error) {
	// Use path as committed (forward slashes).
	spec := "HEAD:" + filepathToGitPath(rel)
	cmd := exec.Command("git", "cat-file", "-p", spec)
	cmd.Dir = root
	out, err := cmd.Output()
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok && len(ee.Stderr) > 0 {
			return nil, fmt.Errorf("git cat-file -p %s: %s", spec, strings.TrimSpace(string(ee.Stderr)))
		}
		return nil, fmt.Errorf("git cat-file -p %s: %w (blob missing on HEAD?)", spec, err)
	}
	return out, nil
}

func filepathToGitPath(rel string) string {
	return strings.ReplaceAll(rel, "\\", "/")
}

func gitRoot() (string, error) {
	out, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
	if err != nil {
		return "", fmt.Errorf("git rev-parse --show-toplevel: %w", err)
	}
	root := strings.TrimSpace(string(out))
	if root == "" {
		return "", fmt.Errorf("empty git toplevel")
	}
	return root, nil
}
