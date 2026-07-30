package browseragent

import (
	"archive/zip"
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
	// Keep placeholder.txt for git/embed safety (do not delete when staging fat xpi).

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
		if v, perr := PeekSignedXPIVersion(absSrc); perr == nil && v != "" {
			version = v
		} else {
			version = inferVersionFromXPIName(filepath.Base(absSrc))
		}
	}
	if version != "" {
		if err := os.WriteFile(filepath.Join(destDir, firefoxXPIVersionFile), []byte(version+"\n"), 0o644); err != nil {
			return "", err
		}
	}
	// Always re-ensure placeholders so install/bundle never leaves tracked files deleted.
	if err := EnsureEmbedPlaceholders(absRoot); err != nil {
		return "", err
	}
	absDest, err := filepath.Abs(destDir)
	if err != nil {
		return "", err
	}
	return absDest, nil
}

// ReadProductVersion reads root VERSION.txt (product SSoT). Empty/missing → "".
func ReadProductVersion(root string) string {
	root = strings.TrimSpace(root)
	if root == "" {
		return ""
	}
	data, err := os.ReadFile(filepath.Join(root, "VERSION.txt"))
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

// PeekSignedXPIVersion returns the extension version for a signed/unsigned .xpi.
// Prefers manifest.json inside the zip; falls back to the file name.
func PeekSignedXPIVersion(xpiPath string) (string, error) {
	xpiPath = strings.TrimSpace(xpiPath)
	if xpiPath == "" {
		return "", fmt.Errorf("xpi path is empty")
	}
	zr, err := zip.OpenReader(xpiPath)
	if err != nil {
		// Still try filename (e.g. corrupt zip).
		if v := inferVersionFromXPIName(filepath.Base(xpiPath)); v != "" {
			return v, nil
		}
		return "", fmt.Errorf("open xpi: %w", err)
	}
	defer zr.Close()

	var topLevel, nested string
	for _, f := range zr.File {
		name := strings.TrimPrefix(f.Name, "./")
		base := filepath.Base(name)
		if base != "manifest.json" {
			continue
		}
		rc, oerr := f.Open()
		if oerr != nil {
			continue
		}
		raw, rerr := io.ReadAll(io.LimitReader(rc, 1<<20))
		_ = rc.Close()
		if rerr != nil {
			continue
		}
		var mani struct {
			Version string `json:"version"`
		}
		if json.Unmarshal(raw, &mani) != nil {
			continue
		}
		v := strings.TrimSpace(mani.Version)
		if v == "" {
			continue
		}
		// Prefer package-root manifest.json over nested paths.
		if name == "manifest.json" || !strings.Contains(name, "/") {
			topLevel = v
			break
		}
		if nested == "" {
			nested = v
		}
	}
	if topLevel != "" {
		return topLevel, nil
	}
	if nested != "" {
		return nested, nil
	}
	if v := inferVersionFromXPIName(filepath.Base(xpiPath)); v != "" {
		return v, nil
	}
	return "", fmt.Errorf("no version in xpi %s", xpiPath)
}

// SignedXPIMatchesProduct reports whether xpiPath's version equals wantVer.
func SignedXPIMatchesProduct(xpiPath, wantVer string) bool {
	wantVer = strings.TrimSpace(wantVer)
	if wantVer == "" {
		return false
	}
	got, err := PeekSignedXPIVersion(xpiPath)
	if err != nil {
		return false
	}
	return strings.TrimSpace(got) == wantVer
}

// IsRealFirefoxXPIFile reports whether path is a non-empty ZIP (.xpi) file.
func IsRealFirefoxXPIFile(path string) bool {
	st, err := os.Stat(path)
	if err != nil || st.IsDir() || st.Size() < 4 {
		return false
	}
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()
	var head [2]byte
	if _, err := io.ReadFull(f, head[:]); err != nil {
		return false
	}
	return head[0] == 'P' && head[1] == 'K'
}

// FirefoxSignHint is the operator command to produce an AMO-signed .xpi.
// Signing can take a long time (Mozilla queue/audit may take hours).
const FirefoxSignHint = "go run ./script/browser-agent/firefox/sign"

// IsAMOSignedXPI reports whether path is a zip that carries Mozilla AMO signing
// artifacts (META-INF/mozilla.rsa and/or mozilla.sf). Filename *signed* alone is
// not sufficient — unsigned local zips fail this check.
func IsAMOSignedXPI(path string) bool {
	if !IsRealFirefoxXPIFile(path) {
		return false
	}
	zr, err := zip.OpenReader(path)
	if err != nil {
		return false
	}
	defer zr.Close()
	var hasRSA, hasSF bool
	for _, f := range zr.File {
		name := strings.ToLower(strings.TrimPrefix(f.Name, "./"))
		name = strings.ReplaceAll(name, "\\", "/")
		switch name {
		case "meta-inf/mozilla.rsa":
			hasRSA = true
		case "meta-inf/mozilla.sf":
			hasSF = true
		}
		if hasRSA && hasSF {
			return true
		}
	}
	// Prefer both; accept rsa alone (some packages still ship rsa without sf naming variants).
	return hasRSA
}

// ValidateReleaseFirefoxXPI checks that path is a real zip, AMO-signed, and
// package version equals wantVer. Returns a detailed error suitable for release.
func ValidateReleaseFirefoxXPI(path, wantVer string) error {
	wantVer = strings.TrimSpace(wantVer)
	if wantVer == "" {
		return fmt.Errorf("product version is empty (set VERSION.txt)")
	}
	if !IsRealFirefoxXPIFile(path) {
		return fmt.Errorf("not a valid .xpi zip: %s", path)
	}
	if !IsAMOSignedXPI(path) {
		return fmt.Errorf("Firefox .xpi exists but is not AMO-signed (missing META-INF/mozilla.rsa)\n  path: %s\n  run: %s\n  note: Mozilla signing/audit may take hours — do not start release until sign finishes",
			path, FirefoxSignHint)
	}
	got, err := PeekSignedXPIVersion(path)
	if err != nil {
		return fmt.Errorf("read xpi version: %w\n  path: %s", err, path)
	}
	if strings.TrimSpace(got) != wantVer {
		return fmt.Errorf("Firefox .xpi version %s != product VERSION.txt %s\n  path: %s\n  run: %s  # after VERSION.txt is %s (AMO may take hours)",
			got, wantVer, path, FirefoxSignHint, wantVer)
	}
	return nil
}

// FindReleaseFirefoxXPI returns an absolute path to an AMO-signed .xpi whose
// package version matches wantVer. Prefers dist/signed, then embed browser-agent.xpi.
func FindReleaseFirefoxXPI(root, wantVer string) (string, error) {
	wantVer = strings.TrimSpace(wantVer)
	if wantVer == "" {
		return "", fmt.Errorf("product version is empty (set VERSION.txt)")
	}
	root = strings.TrimSpace(root)
	if root == "" {
		return "", fmt.Errorf("root is required")
	}

	var tried []string
	signedDir := filepath.Join(root, "dist", "signed")
	if src, err := FindSignedXPIMatchingVersion(signedDir, wantVer); err == nil {
		tried = append(tried, src)
		if err := ValidateReleaseFirefoxXPI(src, wantVer); err == nil {
			abs, _ := filepath.Abs(src)
			return abs, nil
		}
	}

	embedXPI := filepath.Join(root, "browseragent", "embedded", "firefox-xpi", firefoxXPIFileName)
	if IsRealFirefoxXPIFile(embedXPI) {
		tried = append(tried, embedXPI)
		if err := ValidateReleaseFirefoxXPI(embedXPI, wantVer); err == nil {
			abs, _ := filepath.Abs(embedXPI)
			return abs, nil
		}
	}

	// Prefer the most specific error from the last candidate when possible.
	if len(tried) > 0 {
		last := tried[len(tried)-1]
		if err := ValidateReleaseFirefoxXPI(last, wantVer); err != nil {
			return "", err
		}
	}
	return "", fmt.Errorf("release requires an AMO-signed Firefox .xpi for product version %s\n  not found under dist/signed/ or browseragent/embedded/firefox-xpi/\n  sign first (AMO can take a long time — hours):\n    %s\n  then re-run release",
		wantVer, FirefoxSignHint)
}

// EmbedPlaceholderRels are tracked git placeholders under browseragent/embedded.
// Fat staging must not delete these (//go:embed + clean-clone safety).
var EmbedPlaceholderRels = []string{
	"browseragent/embedded/extension/placeholder.txt",
	"browseragent/embedded/session-page/placeholder.txt",
	"browseragent/embedded/firefox-xpi/placeholder.txt",
}

// EnsureEmbedPlaceholders creates empty placeholder.txt files under the three
// //go:embed roots if missing. Safe to call after bundle/stage (does not remove
// fat assets). Prefer keeping placeholders even when real payloads exist.
func EnsureEmbedPlaceholders(root string) error {
	root = strings.TrimSpace(root)
	if root == "" {
		return fmt.Errorf("root is required")
	}
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return err
	}
	for _, rel := range EmbedPlaceholderRels {
		abs := filepath.Join(absRoot, filepath.FromSlash(rel))
		if st, err := os.Stat(abs); err == nil {
			if st.IsDir() {
				return fmt.Errorf("%s: is a directory", rel)
			}
			continue
		} else if !os.IsNotExist(err) {
			return fmt.Errorf("%s: %w", rel, err)
		}
		if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
			return err
		}
		// Empty file matches pre-commit helper; content is irrelevant to runtime.
		if err := os.WriteFile(abs, nil, 0o644); err != nil {
			return fmt.Errorf("write %s: %w", rel, err)
		}
	}
	return nil
}

// RemoveEmbeddedFirefoxXPIIfMismatch deletes embed browser-agent.xpi when its
// package version != wantVer so go install does not ship a stale AMO build.
// Always re-ensures placeholders. No-op if no real xpi or versions match.
func RemoveEmbeddedFirefoxXPIIfMismatch(root, wantVer string) error {
	wantVer = strings.TrimSpace(wantVer)
	embedXPI := filepath.Join(root, "browseragent", "embedded", "firefox-xpi", firefoxXPIFileName)
	if !IsRealFirefoxXPIFile(embedXPI) {
		return EnsureEmbedPlaceholders(root)
	}
	if wantVer != "" && SignedXPIMatchesProduct(embedXPI, wantVer) {
		return EnsureEmbedPlaceholders(root)
	}
	_ = os.Remove(embedXPI)
	_ = os.Remove(filepath.Join(root, "browseragent", "embedded", "firefox-xpi", firefoxXPIVersionFile))
	if err := EnsureEmbedPlaceholders(root); err != nil {
		return err
	}
	return EnsureFirefoxXPIPlaceholder(root)
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
//
// Warning: when multiple versions sit in the same directory, "largest *signed*"
// may pick an older package. Prefer FindSignedXPIMatchingVersion when a product
// version is known.
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

// FindSignedXPIMatchingVersion returns an .xpi under dir whose package version
// (manifest.json inside the zip) equals wantVer. Among matches, prefers names
// containing "signed", then newest ModTime, then largest size.
//
// This avoids picking an older browser-agent-1.0.4-signed.xpi after a successful
// web-ext download of …-1.0.6.xpi in the same directory.
func FindSignedXPIMatchingVersion(dir, wantVer string) (string, error) {
	wantVer = strings.TrimSpace(wantVer)
	if wantVer == "" {
		return "", fmt.Errorf("wantVer is empty")
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", err
	}
	type cand struct {
		path   string
		signed bool
		mtime  int64
		size   int64
	}
	var matches []cand
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if !strings.HasSuffix(strings.ToLower(name), ".xpi") {
			continue
		}
		path := filepath.Join(dir, name)
		if !IsRealFirefoxXPIFile(path) {
			continue
		}
		if !SignedXPIMatchesProduct(path, wantVer) {
			continue
		}
		info, ierr := e.Info()
		if ierr != nil {
			continue
		}
		matches = append(matches, cand{
			path:   path,
			signed: strings.Contains(strings.ToLower(name), "signed") || IsAMOSignedXPI(path),
			mtime:  info.ModTime().UnixNano(),
			size:   info.Size(),
		})
	}
	if len(matches) == 0 {
		return "", fmt.Errorf("no .xpi with version %s under %s", wantVer, dir)
	}
	best := matches[0]
	for _, m := range matches[1:] {
		switch {
		case m.signed && !best.signed:
			best = m
		case m.signed == best.signed && m.mtime > best.mtime:
			best = m
		case m.signed == best.signed && m.mtime == best.mtime && m.size > best.size:
			best = m
		}
	}
	return best.path, nil
}

// EnsureFirefoxXPIPlaceholder writes a compile-safe placeholder when no signed
// xpi is staged (clean clone / --fixture). //go:embed needs at least one file.
// Always keeps/creates placeholder.txt even when a real xpi is present.
func EnsureFirefoxXPIPlaceholder(root string) error {
	if err := EnsureEmbedPlaceholders(root); err != nil {
		return err
	}
	dir := filepath.Join(root, "browseragent", "embedded", "firefox-xpi")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	xpi := filepath.Join(dir, firefoxXPIFileName)
	if IsRealFirefoxXPIFile(xpi) {
		return nil
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
