// Generate syncs the product version from root VERSION.txt into all derived
// places (browseragent embed version, extension manifests, JS fallbacks).
//
//	go run ./script/generate
//	go run ./script/generate --check              # fail if any target would change
//	go run ./script/generate --git-add-generated  # after sync, git add only sinks
//
// Root VERSION.txt is the single source of truth. Run this before install/bundle.
// Install/bundle must call plain generate (no --git-add-generated) so they do not
// mutate the git index.
package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/xhd2015/browser-agent/script/internal/clicolor"
)

// GeneratedVersionSinks is the allowlist of paths stamped from root VERSION.txt.
// --git-add-generated only stages these (when present). Not SSoT VERSION.txt,
// not build/, not embedded fat assets.
var GeneratedVersionSinks = []string{
	"browseragent/VERSION.txt",
	"Chrome-Ext-Browser-Agent/public/manifest.json",
	"Chrome-Ext-Browser-Agent/public/contentScript.js",
	"Chrome-Ext-Browser-Agent/public/background.js",
	"Firefox-Ext-Browser-Agent/public/manifest.json",
	"Firefox-Ext-Browser-Agent/public/contentScript.js",
	"Firefox-Ext-Browser-Agent/public/background.js",
	"Chrome-Ext-Capture-API/public/manifest.json",
}

func main() {
	mode, remain, err := clicolor.ParseFlags(os.Args[1:])
	c := clicolor.NewStyle(mode)
	if err != nil {
		c.ErrorLine(os.Stderr, err.Error())
		os.Exit(1)
	}
	if err := handle(remain, c); err != nil {
		c.ErrorLine(os.Stderr, err.Error())
		os.Exit(1)
	}
}

func handle(args []string, c clicolor.Style) error {
	checkOnly := false
	gitAddGenerated := false
	for _, a := range args {
		switch a {
		case "-h", "--help":
			fmt.Print(helpText())
			return nil
		case "--check":
			checkOnly = true
		case "--git-add-generated":
			gitAddGenerated = true
		default:
			if strings.HasPrefix(a, "-") {
				return fmt.Errorf("unrecognized flag %s (try --help)", a)
			}
			return fmt.Errorf("unexpected argument %q (try --help)", a)
		}
	}
	if checkOnly && gitAddGenerated {
		return fmt.Errorf("cannot combine --check and --git-add-generated")
	}

	root, err := findModuleRoot()
	if err != nil {
		return err
	}

	ver, err := readRootVersion(root)
	if err != nil {
		return err
	}
	fmt.Printf("%s %s %s\n", c.Gray("generate:"), c.Gray("version"), c.Green(ver))

	var dirty []string
	// Paths that exist and are part of the sink set (for git-add allowlist).
	var existingSinks []string
	seen := map[string]bool{}

	noteSink := func(rel string) {
		if seen[rel] {
			return
		}
		path := filepath.Join(root, filepath.FromSlash(rel))
		if st, err := os.Stat(path); err != nil || st.IsDir() {
			return
		}
		seen[rel] = true
		existingSinks = append(existingSinks, rel)
	}

	writeOrCheck := func(rel string, content []byte) error {
		path := filepath.Join(root, filepath.FromSlash(rel))
		existing, err := os.ReadFile(path)
		if err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("read %s: %w", rel, err)
		}
		if bytes.Equal(existing, content) {
			fmt.Printf("  %s  %s\n", c.Gray("ok"), c.Gray(rel))
			noteSink(rel)
			return nil
		}
		if checkOnly {
			dirty = append(dirty, rel)
			fmt.Printf("  %s  %s\n", c.Yellow("dirty"), rel)
			return nil
		}
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(path, content, 0o644); err != nil {
			return fmt.Errorf("write %s: %w", rel, err)
		}
		fmt.Printf("  %s  %s\n", c.Green("wrote"), rel)
		noteSink(rel)
		return nil
	}

	// browseragent/VERSION.txt for //go:embed ClientVersion()
	if err := writeOrCheck("browseragent/VERSION.txt", []byte(ver+"\n")); err != nil {
		return err
	}

	manifests := []string{
		"Chrome-Ext-Browser-Agent/public/manifest.json",
		"Firefox-Ext-Browser-Agent/public/manifest.json",
		"Chrome-Ext-Capture-API/public/manifest.json",
	}
	for _, rel := range manifests {
		path := filepath.Join(root, filepath.FromSlash(rel))
		if _, err := os.Stat(path); err != nil {
			fmt.Printf("  %s  %s %s\n", c.Gray("skip"), c.Gray(rel), c.Gray("(missing)"))
			continue
		}
		updated, err := stampManifestVersion(path, ver)
		if err != nil {
			return fmt.Errorf("%s: %w", rel, err)
		}
		if err := writeOrCheck(rel, updated); err != nil {
			return err
		}
	}

	// JS: contentScript version field + background EXT_VERSION fallback string.
	jsFiles := []struct {
		rel  string
		mode string // "contentscript" | "background"
	}{
		{"Chrome-Ext-Browser-Agent/public/contentScript.js", "contentscript"},
		{"Chrome-Ext-Browser-Agent/public/background.js", "background"},
		{"Firefox-Ext-Browser-Agent/public/contentScript.js", "contentscript"},
		{"Firefox-Ext-Browser-Agent/public/background.js", "background"},
	}
	for _, jf := range jsFiles {
		path := filepath.Join(root, filepath.FromSlash(jf.rel))
		if _, err := os.Stat(path); err != nil {
			fmt.Printf("  %s  %s %s\n", c.Gray("skip"), c.Gray(jf.rel), c.Gray("(missing)"))
			continue
		}
		raw, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		var updated []byte
		switch jf.mode {
		case "contentscript":
			updated = stampContentScriptVersion(raw, ver)
		case "background":
			updated = stampBackgroundFallbackVersion(raw, ver)
		}
		if err := writeOrCheck(jf.rel, updated); err != nil {
			return err
		}
	}

	if checkOnly && len(dirty) > 0 {
		return fmt.Errorf("generate --check: %d file(s) out of date (run: go run ./script/generate):\n  %s",
			len(dirty), strings.Join(dirty, "\n  "))
	}

	if gitAddGenerated {
		// Always stage the full allowlist of existing sinks (even if already "ok"),
		// so unstaged but correct stamps still enter the index for the commit.
		toAdd := existingGeneratedSinks(root)
		if len(toAdd) == 0 {
			fmt.Printf("  %s %s\n", c.Gray("git-add"), c.Gray("0 generated path(s)"))
			return nil
		}
		if err := gitAdd(root, toAdd); err != nil {
			return err
		}
		fmt.Printf("  %s %s\n", c.Gray("git-add"), c.Green(fmt.Sprintf("%d generated path(s)", len(toAdd))))
	}
	return nil
}

func helpText() string {
	var b strings.Builder
	b.WriteString(`Usage: go run ./script/generate [options]

Sync root VERSION.txt into derived version sinks:
`)
	for _, p := range GeneratedVersionSinks {
		b.WriteString("  ")
		b.WriteString(p)
		b.WriteByte('\n')
	}
	b.WriteString(`
Options:
  --check               Fail if any target differs from VERSION.txt (no writes)
  --git-add-generated   After sync, git add only the generated sink paths above
`)
	b.WriteString(clicolor.FlagHelp)
	b.WriteString(`  -h, --help            Show this help

Run from the module root (or any subdirectory). Install/bundle scripts invoke
generate without --git-add-generated so they do not touch the git index.
Pre-commit uses --git-add-generated so version sinks stay in sync on commit.
`)
	return b.String()
}

// existingGeneratedSinks returns allowlisted sink rel paths that exist as files.
func existingGeneratedSinks(root string) []string {
	var out []string
	for _, rel := range GeneratedVersionSinks {
		path := filepath.Join(root, filepath.FromSlash(rel))
		st, err := os.Stat(path)
		if err != nil || st.IsDir() {
			continue
		}
		out = append(out, rel)
	}
	return out
}

// gitAdd runs `git add -- <paths>` at root. Paths must be allowlisted by the caller.
func gitAdd(root string, relPaths []string) error {
	if len(relPaths) == 0 {
		return nil
	}
	args := append([]string{"add", "--"}, relPaths...)
	cmd := exec.Command("git", args...)
	cmd.Dir = root
	out, err := cmd.CombinedOutput()
	if err != nil {
		msg := strings.TrimSpace(string(out))
		if msg != "" {
			return fmt.Errorf("git add: %s", msg)
		}
		return fmt.Errorf("git add: %w", err)
	}
	return nil
}

func readRootVersion(root string) (string, error) {
	path := filepath.Join(root, "VERSION.txt")
	b, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read VERSION.txt: %w", err)
	}
	ver := strings.TrimSpace(string(b))
	// Allow optional leading "v" in SoT; normalize for manifests / ClientVersion.
	ver = strings.TrimPrefix(ver, "v")
	ver = strings.TrimPrefix(ver, "V")
	ver = strings.TrimSpace(ver)
	if ver == "" {
		return "", fmt.Errorf("VERSION.txt is empty")
	}
	// Basic safety: no path separators or spaces.
	if strings.ContainsAny(ver, "/\\ \t\n\r") {
		return "", fmt.Errorf("VERSION.txt has invalid version %q", ver)
	}
	return ver, nil
}

// stampManifestVersion rewrites only the top-level "version" string field so
// existing key order / formatting (including raw <all_urls>) is preserved.
var reManifestVersion = regexp.MustCompile(`(?m)^(\s*"version"\s*:\s*)"[^"]*"`)

func stampManifestVersion(path, ver string) ([]byte, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if !reManifestVersion.Match(raw) {
		return nil, fmt.Errorf("no top-level \"version\" field found")
	}
	out := reManifestVersion.ReplaceAll(raw, []byte(fmt.Sprintf(`${1}"%s"`, ver)))
	// Ensure trailing newline.
	if len(out) == 0 || out[len(out)-1] != '\n' {
		out = append(out, '\n')
	}
	return out, nil
}

// stampContentScriptVersion updates version: "…" inside the marker object.
var reContentScriptVersion = regexp.MustCompile(`(version:\s*)["'][^"']*["']`)

func stampContentScriptVersion(src []byte, ver string) []byte {
	// Prefer the property form used in __BROWSER_AGENT_EXT__ / similar.
	return reContentScriptVersion.ReplaceAll(src, []byte(fmt.Sprintf(`${1}"%s"`, ver)))
}

// stampBackgroundFallbackVersion updates the ternary fallback after BUNDLE_VERSION.
// Matches: : "1.0.1"; following BROWSER_AGENT_BUNDLE_VERSION assignment.
var reBackgroundFallback = regexp.MustCompile(
	`(BROWSER_AGENT_BUNDLE_VERSION[\s\S]{0,200}?:\s*)["'][^"']*["'](\s*;)`,
)

func stampBackgroundFallbackVersion(src []byte, ver string) []byte {
	// Prefer fallback after BROWSER_AGENT_BUNDLE_VERSION (Chrome + Firefox agent SW).
	if reBackgroundFallback.Match(src) {
		return reBackgroundFallback.ReplaceAll(src, []byte(fmt.Sprintf(`${1}"%s"${2}`, ver)))
	}
	// Alternate: EXT_VERSION = … : "x.y.z"
	reAlt := regexp.MustCompile(`(EXT_VERSION\s*=[\s\S]{0,300}?:\s*)["'][^"']*["']`)
	return reAlt.ReplaceAll(src, []byte(fmt.Sprintf(`${1}"%s"`, ver)))
}

func findModuleRoot() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("resolve working directory: %w", err)
	}
	dir := wd
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			if _, err := os.Stat(filepath.Join(dir, "VERSION.txt")); err != nil {
				return "", fmt.Errorf("%s: VERSION.txt not found (not module root?)", dir)
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
