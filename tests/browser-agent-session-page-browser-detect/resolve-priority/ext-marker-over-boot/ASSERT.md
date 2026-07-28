## Expected

Requirement **P2** (`__BROWSER_AGENT_EXT__.browser` over chrome boot):

- Source defines **`resolveInstallBrowser`**.
- Body reads **`__BROWSER_AGENT_EXT__`** and its **`browser`** field.
- When EXT browser is firefox, returns `"firefox"`.
- **Source order**: `__BROWSER_AGENT_EXT__` check appears **before** chrome-boot
  sources (`#browser-agent-boot`, `window.__BROWSER_AGENT` boot stamp, and ideally
  before generic snap path). Content-script marker must not be omitted in favor of
  boot-only detection.
- Documented intent: EXT firefox wins even when boot/meta stamped chrome.

## Side Effects

- None (read-only FS).

## Errors

- Missing `__BROWSER_AGENT_EXT__`, or EXT only after/never vs boot-only path fails.

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

	if !hasContentScriptExtMarker(probe) {
		t.Fatalf("resolveInstallBrowser must read __BROWSER_AGENT_EXT__ (content-script marker); path=%v body=%s",
			resp.FoundPaths, truncate(probe, 800))
	}

	// Must look at .browser on the EXT object (not only the global name).
	if !strings.Contains(probe, "browser") {
		t.Fatalf("resolveInstallBrowser must inspect __BROWSER_AGENT_EXT__.browser; path=%v body=%s",
			resp.FoundPaths, truncate(probe, 600))
	}

	idxExt := strings.Index(probe, "__BROWSER_AGENT_EXT__")
	// Boot stamp markers (not the EXT global). Prefer browser-agent-boot / __BROWSER_AGENT without _EXT_.
	idxBoot := indexOfFirst(probe, "browser-agent-boot", "__BROWSER_AGENT?")
	// Also find bare __BROWSER_AGENT that is not EXT.
	if idxBoot < 0 {
		// Scan for __BROWSER_AGENT not followed by _EXT__
		search := probe
		offset := 0
		for {
			i := strings.Index(search, "__BROWSER_AGENT")
			if i < 0 {
				break
			}
			abs := offset + i
			rest := probe[abs:]
			if strings.HasPrefix(rest, "__BROWSER_AGENT_EXT__") {
				search = search[i+1:]
				offset = abs + 1
				continue
			}
			idxBoot = abs
			break
		}
	}

	if idxExt < 0 {
		t.Fatalf("missing __BROWSER_AGENT_EXT__ index; body=%s", truncate(probe, 600))
	}
	// If boot is also checked, EXT must come first (priority over chrome-stamped boot).
	if idxBoot >= 0 && idxExt > idxBoot {
		t.Fatalf("__BROWSER_AGENT_EXT__ must be checked before boot stamp (idxExt=%d idxBoot=%d); body=%s",
			idxExt, idxBoot, truncate(probe, 900))
	}

	// Positive firefox return path must exist in resolver.
	if !(strings.Contains(probe, `"firefox"`) || strings.Contains(probe, `'firefox'`)) {
		t.Fatalf("resolveInstallBrowser must return firefox when EXT.browser is firefox; path=%v body=%s",
			resp.FoundPaths, truncate(probe, 600))
	}
}
```
