## Expected

Requirement **A3**:

- Final body is HTTP 200 HTML (direct `/` or after one redirect to `/go`).
- Body contains product id `browser-agent` (case-insensitive).
- Prefer also `43761` (soft-fail not required if product alone present, but
  assert both product + port for consistency with product shell).

## Side Effects

- None beyond short-lived control server.

## Errors

- 404-only root without product body fails.

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
		t.Fatalf("GET / body empty; redirect=%q probe=%q", resp.RedirectURL, resp.ProbeURL)
	}
	if !strings.Contains(strings.ToLower(body), "browser-agent") {
		t.Fatalf("GET / HTML must mention product browser-agent; body=%s", truncate(body, 600))
	}
	if !strings.Contains(body, "43761") {
		t.Fatalf("GET / HTML should mention control port 43761; body=%s", truncate(body, 600))
	}
}
```
