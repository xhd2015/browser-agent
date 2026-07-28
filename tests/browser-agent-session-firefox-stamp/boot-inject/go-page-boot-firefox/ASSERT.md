## Expected

- `SessionNew` returns nil error.
- `GET /go?session=` returns HTTP 200 with HTML body.
- Boot browser extracted from `#browser-agent-boot` JSON and/or
  `window.__BROWSER_AGENT` is **`firefox`** (not chrome default).
- Body includes browser-agent-boot marker (or equivalent boot inject).

## Side Effects

- injectSessionBoot / FormatSessionBootJSONWithBrowser ran server-side.

## Errors

- Missing boot, browser=chrome after firefox create, or non-200 fails.

## Exit Code

- N/A.

```go
import (
	"net/http"
	"strings"
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	if resp == nil {
		t.Fatal("resp is nil")
	}
	if resp.SessionNewErr != "" {
		t.Fatalf("SessionNewErr = %q", resp.SessionNewErr)
	}
	if resp.GoStatus != http.StatusOK {
		t.Fatalf("GET /go status=%d body=%s", resp.GoStatus, truncate(resp.GoHTML, 400))
	}
	if strings.TrimSpace(resp.GoHTML) == "" {
		t.Fatal("GET /go returned empty body")
	}
	// Boot script or __BROWSER_AGENT should be present.
	low := strings.ToLower(resp.GoHTML)
	if !strings.Contains(low, "browser-agent-boot") && !strings.Contains(resp.GoHTML, "__BROWSER_AGENT") {
		t.Fatalf("expected boot inject markers; body=%s", truncate(resp.GoHTML, 500))
	}
	got := strings.ToLower(strings.TrimSpace(resp.GoBootBrowser))
	if got != "firefox" {
		t.Fatalf("boot browser = %q, want firefox; metaPath=%q path=%q body=%s",
			resp.GoBootBrowser, resp.MetaPath, resp.MetaExtensionInstallPath, truncate(resp.GoHTML, 600))
	}
}
```
