// Generate syncs the product version from root VERSION.txt into all derived
// places (browseragent embed version, extension manifests, JS fallbacks).
//
//	go run ./script/generate
//	go run ./script/generate --check   # fail if any target would change
//
// Root VERSION.txt is the single source of truth. Run this before install/bundle.
package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func main() {
	if err := handle(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func handle(args []string) error {
	checkOnly := false
	for _, a := range args {
		switch a {
		case "-h", "--help":
			fmt.Print(`Usage: go run ./script/generate [options]

Sync root VERSION.txt into derived version sinks:
  browseragent/VERSION.txt
  Chrome-Ext-Browser-Agent/public/manifest.json (+ contentScript/background fallbacks)
  Firefox-Ext-Browser-Agent/public/manifest.json (+ contentScript/background fallbacks)
  Chrome-Ext-Capture-API/public/manifest.json

Options:
  --check     Fail if any target differs from VERSION.txt (no writes)
  -h, --help  Show this help

Run from the module root (or any subdirectory). Install/bundle scripts invoke
generate automatically before other steps.
`)
			return nil
		case "--check":
			checkOnly = true
		default:
			if strings.HasPrefix(a, "-") {
				return fmt.Errorf("unrecognized flag %s (try --help)", a)
			}
			return fmt.Errorf("unexpected argument %q (try --help)", a)
		}
	}

	root, err := findModuleRoot()
	if err != nil {
		return err
	}

	ver, err := readRootVersion(root)
	if err != nil {
		return err
	}
	fmt.Printf("generate: version %s\n", ver)

	var dirty []string
	writeOrCheck := func(rel string, content []byte) error {
		path := filepath.Join(root, filepath.FromSlash(rel))
		existing, err := os.ReadFile(path)
		if err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("read %s: %w", rel, err)
		}
		if bytes.Equal(existing, content) {
			fmt.Printf("  ok      %s\n", rel)
			return nil
		}
		if checkOnly {
			dirty = append(dirty, rel)
			fmt.Printf("  dirty   %s\n", rel)
			return nil
		}
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(path, content, 0o644); err != nil {
			return fmt.Errorf("write %s: %w", rel, err)
		}
		fmt.Printf("  wrote   %s\n", rel)
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
			fmt.Printf("  skip    %s (missing)\n", rel)
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
			fmt.Printf("  skip    %s (missing)\n", jf.rel)
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
