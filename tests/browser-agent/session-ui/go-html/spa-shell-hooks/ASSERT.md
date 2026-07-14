## Expected

Requirement **E3**:

- HTTP 200; body looks like HTML.
- Body contains live session id.
- Body references `/v1/session` (poller or boot config).
- Body has a status/UI root marker, any of:
  - `data-browser-agent-status` / `data-browser-agent-session`
  - `id="browser-agent-status"` / `browser-agent-root`
  - `data-session-id` / `__BROWSER_AGENT` / product boot JSON with session

## Side Effects

- None.

## Errors

- Empty body or pure 404 page fails.

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
		t.Fatalf("HTML missing session id %q; body=%s", wantID, truncate(body, 500))
	}
	if !strings.Contains(body, "/v1/session") {
		t.Fatalf("HTML must reference /v1/session; body=%s", truncate(body, 500))
	}
	low := strings.ToLower(body)
	hasRoot := strings.Contains(low, "data-browser-agent") ||
		strings.Contains(low, "browser-agent-status") ||
		strings.Contains(low, "browser-agent-root") ||
		strings.Contains(low, "data-session-id") ||
		strings.Contains(low, "__browser_agent") ||
		strings.Contains(low, "id=\"status\"") ||
		strings.Contains(low, "id='status'")
	if !hasRoot {
		t.Fatalf("HTML missing session UI root/marker; body=%s", truncate(body, 600))
	}
}
```
