# Expected

Requirement scenario **#6** — GET `/go?session=valid` HTML:

- HTTP 200.
- Content-Type looks like HTML.
- Body contains the live session id string.
- Body contains a stable status UI root marker:
  - `data-browser-trace-status`, **or**
  - `id="browser-trace-status"` / `id='browser-trace-status'`, **or**
  - `id="status"` combined with browser-trace context (prefer explicit markers).
- Body references the poll endpoint `/v1/session` (inline JS or fetch URL).

## Side Effects

- None (HTML-only smoke; no DOM execution).

## Errors

- Must not return empty body or omit the session id.
- Must not omit both the status UI hook and the poll reference (both required).

## Exit Code

- Not asserted.

```go
import (
	"net/http"
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("probe transport error: %v", err)
	}
	assertHTTPStatus(t, resp, http.StatusOK)
	assertHTMLContentType(t, resp)

	body := resp.BodyString
	if body == "" {
		t.Fatal("HTML body is empty")
	}

	wantID := resp.RealSessionID
	if wantID == "" {
		wantID = req.SessionSuffix
	}
	if !strings.Contains(body, wantID) {
		t.Fatalf("HTML body does not contain session id %q; body=%s",
			wantID, truncate(body, 500))
	}

	low := strings.ToLower(body)
	hasStatusRoot := strings.Contains(low, "data-browser-trace-status") ||
		strings.Contains(low, `id="browser-trace-status"`) ||
		strings.Contains(low, `id='browser-trace-status'`) ||
		strings.Contains(low, "id=browser-trace-status")
	if !hasStatusRoot {
		t.Fatalf("HTML missing status UI root (data-browser-trace-status or id=browser-trace-status); body=%s",
			truncate(body, 600))
	}

	if !strings.Contains(body, "/v1/session") {
		t.Fatalf("HTML must reference poll path /v1/session; body=%s", truncate(body, 600))
	}
}
```
