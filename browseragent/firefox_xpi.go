package browseragent

import (
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

const (
	// Canonical signed XPI filename inside the embed and after extract.
	firefoxXPIFileName = "browser-agent.xpi"
	// firefoxXPIVersionFile is optional stamp staged next to the .xpi.
	firefoxXPIVersionFile = "VERSION.txt"
)

// EmbeddedFirefoxXPIAvailable reports whether the binary embeds a real signed
// .xpi (ZIP magic), not only a placeholder.
func EmbeddedFirefoxXPIAvailable() bool {
	data, err := embeddedFirefoxXPI.ReadFile(embeddedFirefoxXPIRoot + "/" + firefoxXPIFileName)
	if err != nil || len(data) < 4 {
		return false
	}
	// ZIP local-file header starts with PK.
	return data[0] == 'P' && data[1] == 'K'
}

// EmbeddedFirefoxXPIVersion returns the staged VERSION.txt (if any), else "".
func EmbeddedFirefoxXPIVersion() string {
	data, err := embeddedFirefoxXPI.ReadFile(embeddedFirefoxXPIRoot + "/" + firefoxXPIVersionFile)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

// EnsureCanonicalFirefoxXPI writes the embedded signed .xpi under
// ~/.browser-agent/extensions/browser-agent-firefox/browser-agent.xpi and
// returns the absolute path plus version (from VERSION.txt or package manifest).
//
// Errors when the binary has no real signed XPI (placeholder-only embed).
func EnsureCanonicalFirefoxXPI() (path, version string, err error) {
	processEnvMu.Lock()
	defer processEnvMu.Unlock()
	return ensureCanonicalFirefoxXPIBody()
}

// EnsureCanonicalFirefoxXPIWithHome extracts under the given home without
// mutating process env (parallel-safe for doctests).
func EnsureCanonicalFirefoxXPIWithHome(home string) (path, version string, err error) {
	processEnvMu.Lock()
	defer processEnvMu.Unlock()
	if strings.TrimSpace(home) == "" {
		return ensureCanonicalFirefoxXPIBody()
	}
	layout := firefoxExtensionInstallLayoutForHome(home)
	if err := os.MkdirAll(layout.BrowserAgentExtensionsDir, 0o755); err != nil {
		return "", "", fmt.Errorf("mkdir browser-agent-firefox extensions dir: %w", err)
	}
	return extractEmbeddedFirefoxXPI(layout.BrowserAgentExtensionsDir)
}

func ensureCanonicalFirefoxXPIBody() (path, version string, err error) {
	layout, err := DefaultFirefoxExtensionInstallLayout()
	if err != nil {
		return "", "", err
	}
	if err := os.MkdirAll(layout.BrowserAgentExtensionsDir, 0o755); err != nil {
		return "", "", fmt.Errorf("mkdir browser-agent-firefox extensions dir: %w", err)
	}
	return extractEmbeddedFirefoxXPI(layout.BrowserAgentExtensionsDir)
}

// extractEmbeddedFirefoxXPI writes browser-agent.xpi (+ optional VERSION.txt)
// under parentDir (typically ~/.browser-agent/extensions/browser-agent-firefox).
func extractEmbeddedFirefoxXPI(parentDir string) (installPath, version string, err error) {
	if strings.TrimSpace(parentDir) == "" {
		return "", "", fmt.Errorf("parentDir is required")
	}
	if !EmbeddedFirefoxXPIAvailable() {
		return "", "", fmt.Errorf("no signed Firefox .xpi embedded (run: go run ./script/browser-agent/firefox/sign, then reinstall)")
	}
	absBase, err := filepath.Abs(parentDir)
	if err != nil {
		return "", "", fmt.Errorf("resolve parentDir: %w", err)
	}
	if err := os.MkdirAll(absBase, 0o755); err != nil {
		return "", "", fmt.Errorf("create firefox xpi dir: %w", err)
	}

	data, err := embeddedFirefoxXPI.ReadFile(embeddedFirefoxXPIRoot + "/" + firefoxXPIFileName)
	if err != nil {
		return "", "", fmt.Errorf("read embedded %s: %w", firefoxXPIFileName, err)
	}
	dest := filepath.Join(absBase, firefoxXPIFileName)
	if err := os.WriteFile(dest, data, 0o644); err != nil {
		return "", "", fmt.Errorf("write %s: %w", dest, err)
	}

	// Best-effort VERSION.txt next to the xpi for operators.
	if verData, rerr := embeddedFirefoxXPI.ReadFile(embeddedFirefoxXPIRoot + "/" + firefoxXPIVersionFile); rerr == nil {
		_ = os.WriteFile(filepath.Join(absBase, firefoxXPIVersionFile), verData, 0o644)
		version = strings.TrimSpace(string(verData))
	}
	if version == "" {
		version = EmbeddedFirefoxXPIVersion()
	}
	if version == "" {
		if v, merr := peekEmbeddedFirefoxExtensionVersion(); merr == nil {
			version = v
		}
	}

	absDest, err := filepath.Abs(dest)
	if err != nil {
		return "", "", err
	}
	return absDest, version, nil
}

func peekEmbeddedFirefoxExtensionVersion() (string, error) {
	maniBytes, err := embeddedFirefoxExtension.ReadFile(embeddedFirefoxExtensionRoot + "/manifest.json")
	if err != nil {
		return "", err
	}
	var mani struct {
		Version string `json:"version"`
	}
	if err := json.Unmarshal(maniBytes, &mani); err != nil {
		return "", err
	}
	return strings.TrimSpace(mani.Version), nil
}

// StageFirefoxXPIEmbed copies a signed .xpi into
// {root}/browseragent/embedded/firefox-xpi/browser-agent.xpi (+ VERSION.txt).
// Used by install / sign scripts so //go:embed picks up a fat binary.
func StageFirefoxXPIEmbed(root, xpiSrc, version string) (embedDir string, err error) {
	root = strings.TrimSpace(root)
	if root == "" {
		return "", fmt.Errorf("root is required")
	}
	xpiSrc = strings.TrimSpace(xpiSrc)
	if xpiSrc == "" {
		return "", fmt.Errorf("xpiSrc is required")
	}
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return "", err
	}
	absSrc, err := filepath.Abs(xpiSrc)
	if err != nil {
		return "", err
	}
	st, err := os.Stat(absSrc)
	if err != nil {
		return "", fmt.Errorf("stat xpi: %w", err)
	}
	if st.IsDir() || st.Size() < 4 {
		return "", fmt.Errorf("xpi is not a regular non-empty file: %s", absSrc)
	}
	// Verify ZIP magic so we never stage a placeholder as "signed".
	head := make([]byte, 4)
	f, err := os.Open(absSrc)
	if err != nil {
		return "", err
	}
	n, _ := io.ReadFull(f, head)
	_ = f.Close()
	if n < 2 || head[0] != 'P' || head[1] != 'K' {
		return "", fmt.Errorf("not a zip/xpi (missing PK magic): %s", absSrc)
	}

	destDir := filepath.Join(absRoot, "browseragent", "embedded", "firefox-xpi")
	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return "", err
	}
	// Drop placeholder so embed is unambiguously complete.
	_ = os.Remove(filepath.Join(destDir, "placeholder.txt"))

	destXPI := filepath.Join(destDir, firefoxXPIFileName)
	data, err := os.ReadFile(absSrc)
	if err != nil {
		return "", err
	}
	if err := os.WriteFile(destXPI, data, 0o644); err != nil {
		return "", err
	}

	version = strings.TrimSpace(version)
	if version == "" {
		version = inferVersionFromXPIName(filepath.Base(absSrc))
	}
	if version != "" {
		if err := os.WriteFile(filepath.Join(destDir, firefoxXPIVersionFile), []byte(version+"\n"), 0o644); err != nil {
			return "", err
		}
	}
	absDest, err := filepath.Abs(destDir)
	if err != nil {
		return "", err
	}
	return absDest, nil
}

func inferVersionFromXPIName(name string) string {
	// browser-agent-0.3.1-signed.xpi | browser-agent-firefox-0.3.1.xpi | hash-0.3.1.xpi
	name = strings.TrimSuffix(name, ".xpi")
	name = strings.TrimSuffix(name, "-signed")
	parts := strings.Split(name, "-")
	for i := len(parts) - 1; i >= 0; i-- {
		p := parts[i]
		if len(p) > 0 && p[0] >= '0' && p[0] <= '9' && strings.Contains(p, ".") {
			return p
		}
	}
	return ""
}

// FindSignedXPIUnder looks for a preferred signed .xpi under dir (e.g. dist/signed).
// Prefers *signed*.xpi, then any .xpi (largest wins if multiple).
func FindSignedXPIUnder(dir string) (string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", err
	}
	var preferred, any string
	var preferredSize, anySize int64
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if !strings.HasSuffix(strings.ToLower(name), ".xpi") {
			continue
		}
		info, ierr := e.Info()
		if ierr != nil {
			continue
		}
		path := filepath.Join(dir, name)
		low := strings.ToLower(name)
		if strings.Contains(low, "signed") {
			if info.Size() >= preferredSize {
				preferred = path
				preferredSize = info.Size()
			}
			continue
		}
		if info.Size() >= anySize {
			any = path
			anySize = info.Size()
		}
	}
	if preferred != "" {
		return preferred, nil
	}
	if any != "" {
		return any, nil
	}
	return "", fmt.Errorf("no .xpi under %s", dir)
}

// EnsureFirefoxXPIPlaceholder writes a compile-safe placeholder when no signed
// xpi is staged (clean clone / --fixture). //go:embed needs at least one file.
func EnsureFirefoxXPIPlaceholder(root string) error {
	dir := filepath.Join(root, "browseragent", "embedded", "firefox-xpi")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	xpi := filepath.Join(dir, firefoxXPIFileName)
	if st, err := os.Stat(xpi); err == nil && !st.IsDir() && st.Size() >= 4 {
		f, oerr := os.Open(xpi)
		if oerr == nil {
			var head [2]byte
			_, _ = io.ReadFull(f, head[:])
			_ = f.Close()
			if head[0] == 'P' && head[1] == 'K' {
				return nil
			}
		}
	}
	ph := filepath.Join(dir, "placeholder.txt")
	if _, err := os.Stat(ph); err == nil {
		return nil
	}
	return os.WriteFile(ph, []byte("incomplete: signed browser-agent.xpi not staged\n"), 0o644)
}

// PathToFileURL converts an absolute filesystem path to a file:// URL.
// Spaces and special characters are percent-encoded. Empty input → "".
func PathToFileURL(absPath string) string {
	absPath = strings.TrimSpace(absPath)
	if absPath == "" {
		return ""
	}
	if !filepath.IsAbs(absPath) {
		if a, err := filepath.Abs(absPath); err == nil {
			absPath = a
		}
	}
	absPath = filepath.Clean(absPath)
	// url.URL with Scheme file and Path uses path-style (forward slashes).
	p := filepath.ToSlash(absPath)
	// Windows: ensure leading slash so file:///C:/… form is produced.
	if runtime.GOOS == "windows" {
		if !strings.HasPrefix(p, "/") {
			p = "/" + p
		}
	}
	u := url.URL{Scheme: "file", Path: p}
	return u.String()
}

// FirefoxXPIHTTPPath is the control-plane route that serves the signed .xpi.
const FirefoxXPIHTTPPath = "/v1/firefox-xpi"

// listEmbeddedFirefoxXPI is used by diagnostics / tests.
func listEmbeddedFirefoxXPI() ([]string, error) {
	var names []string
	err := fs.WalkDir(embeddedFirefoxXPI, embeddedFirefoxXPIRoot, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(embeddedFirefoxXPIRoot, path)
		if err != nil {
			return err
		}
		names = append(names, rel)
		return nil
	})
	return names, err
}
