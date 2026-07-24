package browseragent

import (
	"fmt"
	"strings"
	"io"
	"os"
	"path/filepath"
)

// ExtensionInstallLayout resolves the canonical operator extension install tree.
type ExtensionInstallLayout struct {
	Root                      string
	BrowserAgentExtensionsDir string
}

// DefaultExtensionInstallLayout resolves home →
// ~/.browser-agent/managed-chrome/extensions/browser-agent/.
func DefaultExtensionInstallLayout() (ExtensionInstallLayout, error) {
	layout, err := DefaultManagedChromeLayout()
	if err != nil {
		return ExtensionInstallLayout{}, err
	}
	return ExtensionInstallLayout{
		Root:                      layout.Root,
		BrowserAgentExtensionsDir: filepath.Join(layout.ExtensionsDir, "browser-agent"),
	}, nil
}

// EnsureCanonicalExtension writes the embedded MV3 package under
// …/extensions/browser-agent/{version}/ and returns the absolute install path
// and version. Idempotent for the same embedded version.
//
// Serializes on processEnvMu so parallel HOME isolators cannot interleave.
func EnsureCanonicalExtension() (path, version string, err error) {
	processEnvMu.Lock()
	defer processEnvMu.Unlock()
	return ensureCanonicalExtensionBody()
}

// EnsureCanonicalExtensionWithHome extracts under {home}/.browser-agent/... without
// mutating process env (parallel-safe for doctests; no t.Setenv).
func EnsureCanonicalExtensionWithHome(home string) (path, version string, err error) {
	processEnvMu.Lock()
	defer processEnvMu.Unlock()
	if strings.TrimSpace(home) == "" {
		return ensureCanonicalExtensionBody()
	}
	layout := ExtensionInstallLayout{
		Root:                      filepath.Join(home, ".browser-agent", "managed-chrome"),
		BrowserAgentExtensionsDir: filepath.Join(home, ".browser-agent", "managed-chrome", "extensions", "browser-agent"),
	}
	if err := os.MkdirAll(layout.BrowserAgentExtensionsDir, 0o755); err != nil {
		return "", "", fmt.Errorf("mkdir browser-agent extensions dir: %w", err)
	}
	return extractEmbeddedExtensionDirect(layout.BrowserAgentExtensionsDir)
}

func ensureCanonicalExtensionBody() (path, version string, err error) {
	layout, err := DefaultExtensionInstallLayout()
	if err != nil {
		return "", "", err
	}
	if err := os.MkdirAll(layout.BrowserAgentExtensionsDir, 0o755); err != nil {
		return "", "", fmt.Errorf("mkdir browser-agent extensions dir: %w", err)
	}
	return extractEmbeddedExtensionDirect(layout.BrowserAgentExtensionsDir)
}

// EnsureCanonicalExtensionLocked is an alias for ensureCanonicalExtensionBody
// for callers that already hold processEnvMu.
func EnsureCanonicalExtensionLocked() (path, version string, err error) {
	return ensureCanonicalExtensionBody()
}

// WarnLoadExtensionIgnored writes the Chrome 137+ operator warning to stderr.
func WarnLoadExtensionIgnored(stderr io.Writer, extPath string) {
	if stderr == nil {
		stderr = os.Stderr
	}
	_, _ = fmt.Fprintf(stderr, "browser-agent: warning: Chrome 137+ ignores --load-extension; Load unpacked from %s\n", extPath)
}