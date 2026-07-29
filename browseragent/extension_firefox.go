package browseragent

import (
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// Firefox extension package segment under ~/.browser-agent (not managed-chrome).
// Layout: {home}/.browser-agent/extensions/browser-agent-firefox/{version}/
const firefoxExtensionSegment = "browser-agent-firefox"

// FirefoxExtensionInstallLayout resolves the canonical Firefox operator install tree.
// Unlike Chrome, this is NOT under managed-chrome/.
type FirefoxExtensionInstallLayout struct {
	Root                      string // {home}/.browser-agent
	BrowserAgentExtensionsDir string // {home}/.browser-agent/extensions/browser-agent-firefox
}

// DefaultFirefoxExtensionInstallLayout resolves home →
// ~/.browser-agent/extensions/browser-agent-firefox/.
func DefaultFirefoxExtensionInstallLayout() (FirefoxExtensionInstallLayout, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return FirefoxExtensionInstallLayout{}, fmt.Errorf("resolve home: %w", err)
	}
	return firefoxExtensionInstallLayoutForHome(home), nil
}

func firefoxExtensionInstallLayoutForHome(home string) FirefoxExtensionInstallLayout {
	root := filepath.Join(home, ".browser-agent")
	return FirefoxExtensionInstallLayout{
		Root:                      root,
		BrowserAgentExtensionsDir: filepath.Join(root, "extensions", firefoxExtensionSegment),
	}
}

// EnsureCanonicalFirefoxExtension writes the embedded Firefox MV3 package under
// ~/.browser-agent/extensions/browser-agent-firefox/{version}/ and returns the
// absolute install path and version. Idempotent for the same package version.
//
// Serializes on processEnvMu so parallel HOME isolators cannot interleave.
func EnsureCanonicalFirefoxExtension() (path, version string, err error) {
	processEnvMu.Lock()
	defer processEnvMu.Unlock()
	return ensureCanonicalFirefoxExtensionBody()
}

// EnsureCanonicalFirefoxExtensionWithHome extracts under
// {home}/.browser-agent/extensions/browser-agent-firefox/{version}/ without
// mutating process env (parallel-safe for doctests; no t.Setenv).
func EnsureCanonicalFirefoxExtensionWithHome(home string) (path, version string, err error) {
	processEnvMu.Lock()
	defer processEnvMu.Unlock()
	if strings.TrimSpace(home) == "" {
		return ensureCanonicalFirefoxExtensionBody()
	}
	layout := firefoxExtensionInstallLayoutForHome(home)
	if err := os.MkdirAll(layout.BrowserAgentExtensionsDir, 0o755); err != nil {
		return "", "", fmt.Errorf("mkdir browser-agent-firefox extensions dir: %w", err)
	}
	return extractEmbeddedFirefoxExtensionDirect(layout.BrowserAgentExtensionsDir)
}

func ensureCanonicalFirefoxExtensionBody() (path, version string, err error) {
	layout, err := DefaultFirefoxExtensionInstallLayout()
	if err != nil {
		return "", "", err
	}
	if err := os.MkdirAll(layout.BrowserAgentExtensionsDir, 0o755); err != nil {
		return "", "", fmt.Errorf("mkdir browser-agent-firefox extensions dir: %w", err)
	}
	return extractEmbeddedFirefoxExtensionDirect(layout.BrowserAgentExtensionsDir)
}

// InstallFirefoxExtension extracts the embedded Firefox package to the canonical
// install path and writes install help to w. Does not open Firefox (library API).
// baseDir is ignored (canonical path is always under home). Output ends with a
// trailing newline. Color is auto (off for non-TTY writers such as bytes.Buffer).
func InstallFirefoxExtension(w io.Writer, baseDir string) error {
	processEnvMu.Lock()
	defer processEnvMu.Unlock()
	colors := newServeColor(w, nil, false, false)
	return installFirefoxExtensionBody(w, baseDir, colors, false)
}

// InstallFirefoxExtensionWithHome isolates HOME for the install (parallel-safe).
// Color is auto (off for non-TTY writers). Does not open Firefox.
func InstallFirefoxExtensionWithHome(w io.Writer, baseDir, home string) error {
	env := map[string]string{}
	if home != "" {
		env["HOME"] = home
	}
	return WithProcessEnv(env, func() error {
		colors := newServeColor(w, env, false, false)
		return installFirefoxExtensionBody(w, baseDir, colors, false)
	})
}

// installFirefoxExtensionBody extracts and writes operator help. colors controls
// ANSI styling (green success/path, orange critical step tokens, gray labels,
// yellow warning:). When openXPI is true and a signed .xpi is available, opens
// the .xpi in Firefox (install prompt). Caller must hold processEnvMu
// (directly or via WithProcessEnv).
func installFirefoxExtensionBody(w io.Writer, baseDir string, colors serveColor, openXPI bool) error {
	if w == nil {
		w = io.Discard
	}
	_ = baseDir
	path, version, err := ensureCanonicalFirefoxExtensionBody()
	if err != nil {
		return err
	}

	// Prefer permanent signed .xpi when embedded.
	var xpiPath, xpiVersion string
	if EmbeddedFirefoxXPIAvailable() {
		xpiPath, xpiVersion, err = ensureCanonicalFirefoxXPIBody()
		if err != nil {
			// Soft: still print temporary path if xpi extract fails.
			xpiPath = ""
			colors.writeWarning(w, "signed .xpi extract failed: "+err.Error())
		}
		if xpiVersion != "" {
			version = xpiVersion
		}
	}

	// Success head (green when color on).
	if _, err := fmt.Fprintln(w, colors.green("Firefox extension extracted for browser-agent.")); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(w); err != nil {
		return err
	}

	if xpiPath != "" {
		if _, err := fmt.Fprintf(w, "  %s      %s\n", colors.gray("xpi"), colors.green(xpiPath)); err != nil {
			return err
		}
	}
	// Unpacked package path (temporary add-on / development).
	if _, err := fmt.Fprintf(w, "  %s     %s\n", colors.gray("path"), colors.green(path)); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "  %s  %s\n", colors.gray("version"), version); err != nil {
		return err
	}

	// Permanent install via signed .xpi (open file → Firefox install prompt).
	if xpiPath != "" {
		if _, err := fmt.Fprintln(w); err != nil {
			return err
		}
		if _, err := fmt.Fprintln(w, colors.gray("Permanent install (recommended) — open the signed .xpi in Firefox:")); err != nil {
			return err
		}
		if _, err := fmt.Fprintln(w); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(w, "     %s\n", colors.green(xpiPath)); err != nil {
			return err
		}
		if _, err := fmt.Fprintln(w); err != nil {
			return err
		}
		if openXPI {
			if err := openFirefoxXPI(xpiPath); err != nil {
				colors.writeWarning(w, "could not open .xpi in Firefox: "+err.Error()+" — open the path above manually")
			} else {
				if _, err := fmt.Fprintln(w, colors.green("Opened the .xpi in Firefox — confirm the install prompt.")); err != nil {
					return err
				}
			}
		} else {
			if _, err := fmt.Fprintf(w, "  Open that file in Firefox (or re-run without %s) to install permanently.\n", colors.orange("--no-open")); err != nil {
				return err
			}
		}
	}

	// Temporary add-on steps (fallback / development; still required markers for CLI tests).
	if _, err := fmt.Fprintln(w); err != nil {
		return err
	}
	if xpiPath != "" {
		if _, err := fmt.Fprintln(w, colors.gray("Or load a temporary add-on (unloads when Firefox restarts):")); err != nil {
			return err
		}
	} else {
		if _, err := fmt.Fprintln(w, colors.gray("Install / load the temporary add-on:")); err != nil {
			return err
		}
	}
	if _, err := fmt.Fprintln(w); err != nil {
		return err
	}
	// 1. about:debugging URL is the critical entry point.
	if _, err := fmt.Fprintf(w, "  1. Open %s\n", colors.orange("about:debugging#/runtime/this-firefox")); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "     (or type %s in the address bar → This Firefox)\n", colors.orange("about:debugging")); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "     Do not use %s — Install from File expects a %s package.\n",
		colors.orange("about:addons"), colors.orange(".xpi")); err != nil {
		return err
	}
	// 2. Button label is the critical click target.
	if _, err := fmt.Fprintf(w, "  2. Click %s\n", colors.orange("Load Temporary Add-on…")); err != nil {
		return err
	}
	// 3. Open the folder only (path on next line; do not mention manifest.json here).
	if _, err := fmt.Fprintf(w, "  3. In the file picker, %s:\n", colors.orange("open this folder")); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(w); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "     %s\n", colors.green(path)); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(w); err != nil {
		return err
	}
	// 4. Select the file (not the folder).
	if _, err := fmt.Fprintf(w, "  4. Select the file %s (not the folder), then Open\n", colors.orange("manifest.json")); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "     (Open stays disabled until %s is selected)\n", colors.orange("manifest.json")); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(w); err != nil {
		return err
	}
	// Restart note with warning: prefix (yellow when color on).
	if xpiPath != "" {
		colors.writeWarning(w, "temporary add-ons unload when Firefox restarts — prefer the signed .xpi for a permanent install.")
	} else {
		colors.writeWarning(w, "temporary add-ons unload when Firefox restarts — re-run install-firefox-extension and Load Temporary Add-on after a restart.")
	}
	return nil
}

// extractEmbeddedFirefoxExtensionDirect writes the embedded Firefox MV3 package
// under {parentDir}/{version}/.
func extractEmbeddedFirefoxExtensionDirect(parentDir string) (installPath, version string, err error) {
	if strings.TrimSpace(parentDir) == "" {
		return "", "", fmt.Errorf("parentDir is required")
	}
	absBase, err := filepath.Abs(parentDir)
	if err != nil {
		return "", "", fmt.Errorf("resolve parentDir: %w", err)
	}

	maniBytes, err := embeddedFirefoxExtension.ReadFile(embeddedFirefoxExtensionRoot + "/manifest.json")
	if err != nil {
		return "", "", fmt.Errorf("read embedded firefox manifest.json: %w", err)
	}
	var mani struct {
		Version string `json:"version"`
	}
	if err := json.Unmarshal(maniBytes, &mani); err != nil {
		return "", "", fmt.Errorf("parse embedded firefox manifest.json: %w", err)
	}
	version = strings.TrimSpace(mani.Version)
	if version == "" {
		return "", "", fmt.Errorf("embedded firefox manifest.json has empty version")
	}

	dest := filepath.Join(absBase, version)
	if err := os.MkdirAll(dest, 0o755); err != nil {
		return "", "", fmt.Errorf("create firefox extract dir: %w", err)
	}

	err = fs.WalkDir(embeddedFirefoxExtension, embeddedFirefoxExtensionRoot, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		rel, err := filepath.Rel(embeddedFirefoxExtensionRoot, path)
		if err != nil {
			return err
		}
		if rel == "." {
			return nil
		}
		rel = filepath.FromSlash(rel)
		target := filepath.Join(dest, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		data, err := embeddedFirefoxExtension.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read embedded firefox %s: %w", path, err)
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		return os.WriteFile(target, data, 0o644)
	})
	if err != nil {
		return "", "", fmt.Errorf("extract embedded firefox extension: %w", err)
	}

	absDest, err := filepath.Abs(dest)
	if err != nil {
		return "", "", err
	}
	return absDest, version, nil
}
