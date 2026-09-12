package browseragent

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// AdhocBrowserProfileRoot returns the /tmp root for an adhoc Chrome profile
// tied to sessionID: /tmp/browser-agent-adhoc-<session-id>.
func AdhocBrowserProfileRoot(sessionID string) string {
	id := strings.TrimSpace(sessionID)
	if id == "" {
		id = "unknown"
	}
	return filepath.Join(os.TempDir(), "browser-agent-adhoc-"+id)
}

// AdhocBrowserProfileDataDir returns the Chrome --user-data-dir path under the
// adhoc root: /tmp/browser-agent-adhoc-<session-id>/data.
func AdhocBrowserProfileDataDir(sessionID string) string {
	return filepath.Join(AdhocBrowserProfileRoot(sessionID), "data")
}

// openAdhocChrome launches Chrome with an isolated user-data-dir and
// --load-extension. Does not join the operator's default Chrome profile.
func openAdhocChrome(sessionURL, extensionPath, dataDir string) error {
	sessionURL = strings.TrimSpace(sessionURL)
	if sessionURL == "" {
		return fmt.Errorf("open adhoc chrome: empty session URL")
	}
	dataDir = strings.TrimSpace(dataDir)
	if dataDir == "" {
		return fmt.Errorf("open adhoc chrome: empty data dir")
	}
	extensionPath = strings.TrimSpace(extensionPath)
	if extensionPath == "" {
		return fmt.Errorf("open adhoc chrome: empty extension path")
	}
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		return fmt.Errorf("open adhoc chrome: mkdir data dir: %w", err)
	}
	args := BuildManagedChromeArgs(dataDir, extensionPath, sessionURL)
	return launchChromeWithArgs(args)
}
