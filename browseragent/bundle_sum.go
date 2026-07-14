package browseragent

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// BundleSum is the identity pair stored in / parsed from bundle-sum.js.
type BundleSum struct {
	Version string
	MD5     string // lowercase hex, 32 chars preferred
}

// Extension match statuses for GET /v1/session extension_match.
const (
	ExtensionMatchNotConnected    = "not_connected"
	ExtensionMatchOK              = "ok"
	ExtensionMatchVersionMismatch = "version_mismatch"
	ExtensionMatchMD5Mismatch     = "md5_mismatch"
	ExtensionMatchMD5Unknown      = "md5_unknown"
)

const bundleSumFileName = "bundle-sum.js"

var (
	reBundleVersion = regexp.MustCompile(`(?m)BROWSER_AGENT_BUNDLE_VERSION\s*=\s*["']([^"']*)["']`)
	reBundleMD5     = regexp.MustCompile(`(?m)BROWSER_AGENT_BUNDLE_MD5\s*=\s*["']([^"']*)["']`)
)

// ParseBundleSumJS parses a SW-loadable bundle-sum.js without a JS engine.
func ParseBundleSumJS(data []byte) (BundleSum, error) {
	if len(data) == 0 {
		return BundleSum{}, fmt.Errorf("bundle-sum.js: empty")
	}
	s := string(data)
	vm := reBundleVersion.FindStringSubmatch(s)
	mm := reBundleMD5.FindStringSubmatch(s)
	if vm == nil || mm == nil {
		return BundleSum{}, fmt.Errorf("bundle-sum.js: missing BROWSER_AGENT_BUNDLE_VERSION or BROWSER_AGENT_BUNDLE_MD5")
	}
	version := strings.TrimSpace(vm[1])
	md5hex := strings.ToLower(strings.TrimSpace(mm[1]))
	if version == "" {
		return BundleSum{}, fmt.Errorf("bundle-sum.js: empty version")
	}
	if md5hex == "" {
		return BundleSum{}, fmt.Errorf("bundle-sum.js: empty md5")
	}
	return BundleSum{Version: version, MD5: md5hex}, nil
}

// ReadBundleSumFromDir reads and parses dir/bundle-sum.js.
func ReadBundleSumFromDir(dir string) (BundleSum, error) {
	path := filepath.Join(dir, bundleSumFileName)
	data, err := os.ReadFile(path)
	if err != nil {
		return BundleSum{}, err
	}
	return ParseBundleSumJS(data)
}

// WriteBundleSumJS writes dir/bundle-sum.js with version + md5 assignments.
func WriteBundleSumJS(dir string, version, md5hex string) error {
	if strings.TrimSpace(dir) == "" {
		return fmt.Errorf("WriteBundleSumJS: dir is required")
	}
	version = strings.TrimSpace(version)
	md5hex = strings.ToLower(strings.TrimSpace(md5hex))
	if version == "" {
		return fmt.Errorf("WriteBundleSumJS: version is required")
	}
	if md5hex == "" {
		return fmt.Errorf("WriteBundleSumJS: md5 is required")
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	js := FormatBundleSumJS(version, md5hex)
	return os.WriteFile(filepath.Join(dir, bundleSumFileName), []byte(js), 0o644)
}

// FormatBundleSumJS returns the generated JS body for bundle-sum.js.
func FormatBundleSumJS(version, md5hex string) string {
	return fmt.Sprintf(
		"// browser-agent bundle-sum — generated; do not edit\nvar BROWSER_AGENT_BUNDLE_VERSION = %q;\nvar BROWSER_AGENT_BUNDLE_MD5 = %q;\n",
		version, strings.ToLower(strings.TrimSpace(md5hex)),
	)
}

// ComputeExtensionContentMD5 hashes extension dir contents with sorted relative
// paths. Basename bundle-sum.js is excluded so the sum file does not change the digest.
//
// Digest stream per file (slash-normalized relative path):
//
//	path + "\n" + file contents + "\n"
func ComputeExtensionContentMD5(dir string) (string, error) {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return "", err
	}
	st, err := os.Stat(abs)
	if err != nil {
		return "", err
	}
	if !st.IsDir() {
		return "", fmt.Errorf("ComputeExtensionContentMD5: not a directory: %s", abs)
	}

	var paths []string
	err = filepath.Walk(abs, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if info.IsDir() {
			base := info.Name()
			if strings.HasPrefix(base, ".") && base != "." {
				return filepath.SkipDir
			}
			return nil
		}
		base := info.Name()
		if base == bundleSumFileName {
			return nil
		}
		if strings.HasPrefix(base, ".") {
			return nil
		}
		rel, err := filepath.Rel(abs, path)
		if err != nil {
			return err
		}
		// Normalize to slash paths for cross-platform stability.
		rel = filepath.ToSlash(rel)
		paths = append(paths, rel)
		return nil
	})
	if err != nil {
		return "", err
	}
	sort.Strings(paths)

	h := md5.New()
	for _, rel := range paths {
		data, err := os.ReadFile(filepath.Join(abs, filepath.FromSlash(rel)))
		if err != nil {
			return "", err
		}
		_, _ = h.Write([]byte(rel))
		_, _ = h.Write([]byte("\n"))
		_, _ = h.Write(data)
		_, _ = h.Write([]byte("\n"))
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// EnsureExtensionBundleSum (re)writes dir/bundle-sum.js from the content hash
// of the staged package (excluding the sum file). versionOverride (if non-empty)
// is preferred; otherwise manifest.json version is used.
func EnsureExtensionBundleSum(dir string, versionOverride string) (BundleSum, error) {
	version := strings.TrimSpace(versionOverride)
	if version == "" {
		version = readManifestVersion(dir)
	}
	if version == "" {
		version = "0.0.0"
	}
	md5hex, err := ComputeExtensionContentMD5(dir)
	if err != nil {
		return BundleSum{}, err
	}
	if err := WriteBundleSumJS(dir, version, md5hex); err != nil {
		return BundleSum{}, err
	}
	return BundleSum{Version: version, MD5: md5hex}, nil
}

func readManifestVersion(dir string) string {
	data, err := os.ReadFile(filepath.Join(dir, "manifest.json"))
	if err != nil {
		return ""
	}
	var mani struct {
		Version string `json:"version"`
	}
	if err := json.Unmarshal(data, &mani); err != nil {
		return ""
	}
	return strings.TrimSpace(mani.Version)
}

// ComputeExtensionMatch compares loaded hello identity against the embedded package.
// Priority: not_connected → version_mismatch → md5_unknown → md5_mismatch → ok.
func ComputeExtensionMatch(connected bool, embedded, loaded BundleSum) string {
	if !connected {
		return ExtensionMatchNotConnected
	}
	if strings.TrimSpace(loaded.Version) != strings.TrimSpace(embedded.Version) {
		return ExtensionMatchVersionMismatch
	}
	loadedMD5 := strings.ToLower(strings.TrimSpace(loaded.MD5))
	if loadedMD5 == "" {
		return ExtensionMatchMD5Unknown
	}
	if loadedMD5 != strings.ToLower(strings.TrimSpace(embedded.MD5)) {
		return ExtensionMatchMD5Mismatch
	}
	return ExtensionMatchOK
}

// FormatExtensionMismatchWarning builds a human-readable mismatch warning.
func FormatExtensionMismatchWarning(embedded, loaded BundleSum, installPath string) string {
	return fmt.Sprintf(
		"extension identity mismatch: embedded version=%s md5=%s; loaded version=%s md5=%s; match depends on install path %s",
		embedded.Version, embedded.MD5, loaded.Version, loaded.MD5, installPath,
	)
}

// ColorOrangeIfTTY wraps s with 256-color orange SGR when isTTY; otherwise returns s unchanged.
func ColorOrangeIfTTY(s string, isTTY bool) string {
	if !isTTY {
		return s
	}
	return "\x1b[38;5;208m" + s + "\x1b[0m"
}


