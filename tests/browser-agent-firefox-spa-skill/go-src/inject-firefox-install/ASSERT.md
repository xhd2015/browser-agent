## Expected

Requirement **G1** (Go inject / fallback Firefox install):

- At least one `browseragent/*.go` source is found (prefer `server.go`).
- Combined Go sources that participate in session-page install must document
  Firefox install UX with **all three**:
  1. **`about:debugging`**
  2. temporary add-on / **Load Temporary Add-on**
  3. **`browser-agent-firefox`**
- Prefer markers near `injectSessionBoot`, `writeFallbackSessionHTML`,
  `data-browser-agent-install`, or a dedicated firefox install HTML helper.
- Accept `install-firefox-extension` as additional help text (not a substitute
  for about:debugging + temporary add-on + path segment).

## Side Effects

- None (read-only FS).

## Errors

- Chrome-only inject/fallback (no Firefox branch) fails Phase 4 session UX.

## Exit Code

- Not asserted.

```go
import (
	"path/filepath"
	"strings"
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	src := assertFilePresent(t, req, resp, "browseragent inject/fallback")

	// Prefer session-page inject surface over pure CLI help alone.
	// If markers only appear in extension_firefox.go, still require that
	// server inject/fallback OR a shared HTML helper also participates —
	// checked by ensuring about:debugging appears in a file that mentions
	// injectSessionBoot / writeFallbackSessionHTML / browser-agent-install
	// OR combined sources clearly wire firefox into install HTML.
	sessionSurface := extractSessionInstallSurface(resp)
	probe := sessionSurface
	if strings.TrimSpace(probe) == "" {
		probe = src
	}

	if !hasAboutDebugging(probe) {
		t.Fatalf("session install surface missing about:debugging; paths=%v snippet=%s",
			resp.FoundPaths, truncate(probe, 700))
	}
	if !hasTemporaryAddon(probe) {
		t.Fatalf("session install surface missing temporary add-on wording; paths=%v snippet=%s",
			resp.FoundPaths, truncate(probe, 700))
	}
	if !hasFirefoxExtPath(probe) {
		t.Fatalf("session install surface missing browser-agent-firefox; paths=%v snippet=%s",
			resp.FoundPaths, truncate(probe, 700))
	}

	// Ensure not solely CLI extension_firefox.go without any session-page hook.
	if !sessionInstallHooksPresent(resp) {
		t.Fatalf("Firefox install markers must be reachable from injectSessionBoot / writeFallbackSessionHTML / data-browser-agent-install (not CLI-only); paths=%v",
			resp.FoundPaths)
	}
}

// extractSessionInstallSurface prefers files that define inject/fallback/install panels.
func extractSessionInstallSurface(resp *Response) string {
	if resp == nil || resp.FileContents == nil {
		return ""
	}
	var parts []string
	for p, body := range resp.FileContents {
		base := filepath.Base(p)
		if base == "extension_firefox.go" {
			// CLI help is secondary; include only if also session-page hooks exist elsewhere.
			continue
		}
		if strings.Contains(body, "injectSessionBoot") ||
			strings.Contains(body, "writeFallbackSessionHTML") ||
			strings.Contains(body, "data-browser-agent-install") ||
			strings.Contains(body, "browser-agent-install") ||
			strings.Contains(body, "about:debugging") && strings.Contains(body, "chrome://extensions") {
			parts = append(parts, body)
		}
	}
	if len(parts) == 0 {
		// Fall back: any non-extension_firefox file with about:debugging.
		for p, body := range resp.FileContents {
			if filepath.Base(p) == "extension_firefox.go" {
				continue
			}
			if hasAboutDebugging(body) || hasFirefoxExtPath(body) {
				parts = append(parts, body)
			}
		}
	}
	return strings.Join(parts, "\n")
}

func sessionInstallHooksPresent(resp *Response) bool {
	if resp == nil || resp.FileContents == nil {
		return false
	}
	combined := ""
	for p, body := range resp.FileContents {
		if filepath.Base(p) == "extension_firefox.go" {
			continue
		}
		combined += body + "\n"
	}
	// Session-page hook: inject or fallback knows about firefox install.
	hasHook := strings.Contains(combined, "injectSessionBoot") ||
		strings.Contains(combined, "writeFallbackSessionHTML") ||
		strings.Contains(combined, "data-browser-agent-install")
	if !hasHook {
		return false
	}
	// And that surface (or helpers it can call in same package files excluding pure CLI) has firefox markers.
	return hasAboutDebugging(combined) && hasTemporaryAddon(combined) && hasFirefoxExtPath(combined)
}
```
