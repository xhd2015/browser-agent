## Expected

Requirement **A2**:

- HTTP 200; body looks like HTML.
- Body contains live session id.
- Body references `/v1/session`.
- Body contains `43761`.
- Body contains `browser-agent` (case-insensitive).
- Boot surface present: `#browser-agent-boot` / `data-session-id` /
  `data-control-port` / `__BROWSER_AGENT` / `control_port` / `controlPort`.
- Root mount present (`id="root"` or `data-browser-agent-root`).

## Side Effects

- None beyond short-lived control server.

## Errors

- Empty body / 404 / missing session or product markers fails.

## Exit Code

- Not asserted.

```go
import (
	"net/http"
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	assertHTTPStatus(t, resp, http.StatusOK)
	assertHTMLContentType(t, resp)
	body := resp.BodyString
	if strings.TrimSpace(body) == "" {
		t.Fatal("HTML body empty")
	}
	wantID := resp.RealSessionID
	if wantID == "" {
		wantID = req.SessionID
	}
	if !strings.Contains(body, wantID) {
		t.Fatalf("HTML missing session id %q; body=%s", wantID, truncate(body, 600))
	}
	if !strings.Contains(body, "/v1/session") {
		t.Fatalf("HTML must reference /v1/session; body=%s", truncate(body, 500))
	}
	if !strings.Contains(body, "43761") {
		t.Fatalf("HTML/boot must mention port 43761; body=%s", truncate(body, 600))
	}
	if !strings.Contains(strings.ToLower(body), "browser-agent") {
		t.Fatalf("HTML/boot must mention browser-agent; body=%s", truncate(body, 600))
	}
	if !hasBootOrProductMarkers(body) {
		t.Fatalf("HTML missing boot markers (browser-agent-boot / data-session-id / __BROWSER_AGENT / control_port); body=%s",
			truncate(body, 700))
	}
	if !hasRootMount(body) {
		t.Fatalf("HTML missing root mount; body=%s", truncate(body, 500))
	}
}
```
