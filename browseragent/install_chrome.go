package browseragent

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/xhd2015/browser-agent/browseragent/inject"
	"github.com/xhd2015/dot-pkgs/go-pkgs/computer-use/macos/chrome"
)

// InstallChromeUIConfig controls optional Chrome UI Load unpacked after extract.
type InstallChromeUIConfig struct {
	// OpenUI drives chrome://extensions Load unpacked when true.
	OpenUI bool
	// DryRun opens extensions page and reports controls; no Load unpacked click.
	DryRun bool
	// DumpTree dumps UI tree to Stderr (no load).
	DumpTree bool
	// ExtensionDir overrides the folder loaded by UI (still extracts canonical embed).
	// Empty → use just-extracted path.
	ExtensionDir string
	// KeepOlder skips post-load removal of other same-name extension cards
	// (default: remove older best-effort after successful UI load).
	KeepOlder bool
	// ScreenshotDir when set saves a PNG after each UI step (debug).
	ScreenshotDir string
	// JSONResultPath when set writes open-handoff JSON (ok/path/ui) after extract.
	// CLI --write-json-result; implies no Chrome UI.
	JSONResultPath string
	// Stderr receives warnings and dump-tree (defaults to discarding if nil when no UI).
	Stderr io.Writer
	// Colors styles warning/success lines (optional zero = no color).
	Colors serveColor
}

// installChromeJSONResult is the --write-json-result payload for a GUI host
// (Marcus) to open chrome://extensions and reveal the unpacked folder.
type installChromeJSONResult struct {
	OK   bool                  `json:"ok"`
	Path string                `json:"path"`
	UI   []installChromeJSONUI `json:"ui"`
}

type installChromeJSONUI struct {
	Kind string `json:"kind"`
	App  string `json:"app,omitempty"`
	URL  string `json:"url,omitempty"`
	Path string `json:"path,omitempty"`
}

// InstallChromeExtension extracts the embedded extension to the canonical install
// path and writes user-facing Load unpacked instructions to w. baseDir is ignored
// (canonical path is always under home). Does not drive Chrome UI (package API /
// doctest safe). Output ends with a trailing newline.
func InstallChromeExtension(w io.Writer, baseDir string) error {
	processEnvMu.Lock()
	defer processEnvMu.Unlock()
	return installChromeExtensionBody(w, baseDir, InstallChromeUIConfig{})
}

// InstallChromeExtensionWithHome isolates HOME for the install (parallel-safe).
func InstallChromeExtensionWithHome(w io.Writer, baseDir, home string) error {
	env := map[string]string{}
	if home != "" {
		env["HOME"] = home
	}
	return WithProcessEnv(env, func() error {
		return installChromeExtensionBody(w, baseDir, InstallChromeUIConfig{})
	})
}

// InstallChromeExtensionWithUI extracts then optionally drives Load unpacked UI.
func InstallChromeExtensionWithUI(w io.Writer, baseDir string, ui InstallChromeUIConfig) error {
	processEnvMu.Lock()
	defer processEnvMu.Unlock()
	return installChromeExtensionBody(w, baseDir, ui)
}

func installChromeExtensionBody(w io.Writer, baseDir string, ui InstallChromeUIConfig) error {
	if w == nil {
		w = io.Discard
	}
	_ = baseDir
	path, version, err := ensureCanonicalExtensionBody()
	if err != nil {
		return err
	}
	// Prefer identity from bundle-sum.js (canonical after EnsureExtensionBundleSum).
	sum, sumErr := ReadBundleSumFromDir(path)
	if sumErr == nil && strings.TrimSpace(sum.Version) != "" {
		version = sum.Version
	}
	md5hex := ""
	if sumErr == nil {
		md5hex = sum.MD5
	}
	if md5hex == "" {
		if s, e := EnsureExtensionBundleSum(path, version); e == nil {
			version = s.Version
			md5hex = s.MD5
		}
	}

	loadPath := path
	if strings.TrimSpace(ui.ExtensionDir) != "" {
		loadPath = strings.TrimSpace(ui.ExtensionDir)
	}

	_, err = fmt.Fprintf(w, `Chrome extension extracted for browser-agent.

  path     %s
  version  %s
  md5      %s

Install / load the unpacked extension:

  1. Open chrome://extensions
  2. Enable Developer mode (top-right toggle)
  3. Click Load unpacked
  4. Select this folder:

     %s
`, path, version, md5hex, loadPath)
	if err != nil {
		return err
	}
	if loadPath != path {
		if _, err := fmt.Fprintf(w, "\n  (UI load path override: %s)\n", loadPath); err != nil {
			return err
		}
	}
	if _, err := fmt.Fprint(w, `
After loading, keep the session page open so the extension can connect.
Compare version/md5 with browser-agent serve "embedded" lines if connection warns of mismatch.
`); err != nil {
		return err
	}

	if dest := strings.TrimSpace(ui.JSONResultPath); dest != "" {
		if err := writeInstallChromeJSONResult(dest, loadPath); err != nil {
			return err
		}
	}

	wantUI := ui.OpenUI || ui.DryRun || ui.DumpTree
	if !wantUI {
		return nil
	}

	stderr := ui.Stderr
	if stderr == nil {
		stderr = io.Discard
	}
	colors := ui.Colors

	// Injected test hook (no real Chrome).
	if fn := inject.ChromeLoadFn(); fn != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()
		if err := fn(ctx, loadPath, ui.DryRun, ui.DumpTree); err != nil {
			colors.writeWarning(stderr, "could not load unpacked extension via UI: "+err.Error()+" — use chrome://extensions and the path above")
			return nil
		}
		if ui.DumpTree {
			return nil
		}
		if ui.DryRun {
			if _, err := fmt.Fprintln(w, colors.gray("dry-run complete (UI probe via inject)")); err != nil {
				return err
			}
			return nil
		}
		if _, err := fmt.Fprintln(w, colors.green("Loaded unpacked extension in Chrome (Developer mode → Load unpacked).")); err != nil {
			return err
		}
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	res, err := chrome.LoadUnpacked(ctx, chrome.LoadUnpackedOpts{
		ExtensionDir:  loadPath,
		DryRun:        ui.DryRun,
		DumpTree:      ui.DumpTree,
		VerifyName:    "Browser Agent",
		VerifyVersion: version,
		KeepOlder:     ui.KeepOlder,
		ScreenshotDir: ui.ScreenshotDir,
		Stdout:        w,
		Stderr:        stderr,
	})
	if err != nil {
		if errors.Is(err, chrome.ErrUnsupported) {
			colors.writeWarning(stderr, "automatic Load unpacked UI is only supported on macOS — use the path above manually")
			return nil
		}
		colors.writeWarning(stderr, "could not load unpacked extension via UI: "+err.Error()+" — use chrome://extensions and the path above")
		return nil
	}
	if ui.DumpTree {
		return nil
	}
	if ui.DryRun {
		return nil
	}
	if res.Loaded {
		if _, err := fmt.Fprintln(w, colors.green("Loaded unpacked extension in Chrome (Developer mode → Load unpacked).")); err != nil {
			return err
		}
	} else if res.SubmittedUnknown {
		if _, err := fmt.Fprintln(w, colors.gray("Load unpacked submitted (could not confirm extension card).")); err != nil {
			return err
		}
	}
	if res.RemoveOlderAttempted && res.RemovedOlder > 0 {
		if _, err := fmt.Fprintf(w, "%s\n", colors.gray(fmt.Sprintf("Removed %d older same-name extension card(s).", res.RemovedOlder))); err != nil {
			return err
		}
	}
	return nil
}

func writeInstallChromeJSONResult(dest, loadPath string) error {
	dest = strings.TrimSpace(dest)
	if dest == "" {
		return fmt.Errorf("--write-json-result requires a file path")
	}
	payload := installChromeJSONResult{
		OK:   true,
		Path: loadPath,
		UI: []installChromeJSONUI{
			{Kind: "open-url", App: chrome.DefaultAppName, URL: chrome.ExtensionsURL},
			{Kind: "reveal", Path: loadPath},
		},
	}
	raw, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return fmt.Errorf("write-json-result: %w", err)
	}
	raw = append(raw, '\n')
	if err := os.WriteFile(dest, raw, 0o644); err != nil {
		return fmt.Errorf("write-json-result: %w", err)
	}
	return nil
}
