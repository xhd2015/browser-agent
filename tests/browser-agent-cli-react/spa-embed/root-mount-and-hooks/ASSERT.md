## Expected

Requirement **D1**:

- HTTP 200; body looks like HTML.
- React/root mount present: `id="root"` or `data-browser-agent-root` or
  `browser-agent-root` / `id='root'`.
- Body references `/v1/session`.
- Body contains `43761`.
- Body contains `browser-agent` (case-insensitive).

## Side Effects

- None.

## Errors

- Empty body / pure 404 / wrong product port fails.

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
	low := strings.ToLower(body)
	hasRoot := strings.Contains(low, `id="root"`) ||
		strings.Contains(low, `id='root'`) ||
		strings.Contains(low, "data-browser-agent-root") ||
		strings.Contains(low, "browser-agent-root") ||
		strings.Contains(low, `id="app"`) ||
		strings.Contains(low, "data-reactroot")
	if !hasRoot {
		t.Fatalf("HTML missing React/root mount marker; body=%s", truncate(body, 600))
	}
	if !strings.Contains(body, "/v1/session") {
		t.Fatalf("HTML must reference /v1/session; body=%s", truncate(body, 500))
	}
	if !strings.Contains(body, "43761") {
		t.Fatalf("HTML/boot config must mention product port 43761; body=%s", truncate(body, 600))
	}
	if !strings.Contains(low, "browser-agent") {
		t.Fatalf("HTML/boot config must mention browser-agent; body=%s", truncate(body, 600))
	}
}
```
