## Expected

Requirement **P4** (legacy snap/boot/path fallback kept):

- Source defines **`resolveInstallBrowser`**.
- Still considers **at least two** of:
  - `snap` / `browsers` array / `snap.browser`
  - install path segment **`browser-agent-firefox`**
  - boot: `__BROWSER_AGENT` (non-EXT) and/or `#browser-agent-boot` / `browser-agent-boot`
- These remain valid firefox signals when forced / EXT / FF-UA are absent.
- Must not delete all legacy paths in favor of UA-only.

## Side Effects

- None (read-only FS).

## Errors

- Resolver with zero snap/boot/path firefox signals fails.

## Exit Code

- Not asserted.

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	src := assertSourcePresent(t, req, resp, "SessionPageApp/helpers")

	if !hasResolveInstallBrowser(src) {
		t.Fatalf("resolveInstallBrowser must be defined; path=%v snippet=%s",
			resp.FoundPaths, truncate(src, 500))
	}

	body := extractNamedFuncBody(src, "resolveInstallBrowser")
	probe := body
	if probe == "" {
		probe = src
	}

	score := 0
	if strings.Contains(probe, "browsers") || strings.Contains(probe, "snap") {
		score++
	}
	if strings.Contains(probe, "browser-agent-firefox") || strings.Contains(probe, "installPath") {
		score++
	}
	// Boot without requiring EXT-only.
	hasBoot := strings.Contains(probe, "browser-agent-boot") ||
		strings.Contains(probe, "__BROWSER_AGENT")
	// If only EXT is present, still count as a window/boot-ish surface but prefer real boot.
	if hasBoot {
		score++
	}

	if score < 2 {
		t.Fatalf("resolveInstallBrowser should keep snap/boot/path firefox fallback (score=%d want>=2); path=%v body=%s",
			score, resp.FoundPaths, truncate(probe, 900))
	}

	if !(strings.Contains(probe, `"firefox"`) || strings.Contains(probe, `'firefox'`)) {
		t.Fatalf("fallback path must still be able to return firefox; path=%v body=%s",
			resp.FoundPaths, truncate(probe, 500))
	}
}
```
