## Expected

Requirement **D1** (`detectRuntimeBrowser` Firefox UA):

- Source defines **`detectRuntimeBrowser`** (export preferred; clear name required).
- Helper reads UA via argument and/or `navigator.userAgent` / `userAgent`.
- Detects Firefox with the **`Firefox/`** token (or `/Firefox\//` / equivalent
  regex that matches the Firefox UA product token).
- On match, returns **`"firefox"`** (or `'firefox'` / `` `firefox` ``).

Acceptable shapes:

```ts
function detectRuntimeBrowser(userAgent?: string): InstallBrowser {
  const ua = userAgent ?? navigator.userAgent;
  if (ua.includes("Firefox/") || /Firefox\//i.test(ua)) return "firefox";
  return "chrome";
}
```

## Side Effects

- None (read-only FS).

## Errors

- Missing helper name, missing UA read, or missing Firefox token / firefox return fails.

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

	if !hasDetectRuntimeBrowser(src) {
		t.Fatalf("detectRuntimeBrowser must be defined (Phase 1 pure UA helper); path=%v snippet=%s",
			resp.FoundPaths, truncate(src, 500))
	}

	body := extractNamedFuncBody(src, "detectRuntimeBrowser")
	probe := body
	if probe == "" {
		probe = src
	}

	if !hasUserAgentRead(probe) {
		t.Fatalf("detectRuntimeBrowser must read userAgent / navigator.userAgent; path=%v body=%s",
			resp.FoundPaths, truncate(probe, 600))
	}
	if !hasFirefoxUAToken(probe) {
		t.Fatalf("detectRuntimeBrowser must detect Firefox/ UA token; path=%v body=%s",
			resp.FoundPaths, truncate(probe, 600))
	}

	// Must return firefox for the Firefox branch.
	hasFirefoxReturn := strings.Contains(probe, `"firefox"`) ||
		strings.Contains(probe, `'firefox'`) ||
		strings.Contains(probe, "`firefox`")
	if !hasFirefoxReturn {
		t.Fatalf("detectRuntimeBrowser must return \"firefox\" for Firefox UA; path=%v body=%s",
			resp.FoundPaths, truncate(probe, 600))
	}
}
```
