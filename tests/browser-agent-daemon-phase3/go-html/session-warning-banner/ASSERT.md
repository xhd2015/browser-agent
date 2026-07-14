## Expected

- HTTP status **200**; body is HTML.
- Body contains attribute/marker `data-browser-agent-session-warning`.
- Body contains session id `sess-go`.
- Body references `/v1/session` (poll hook or boot).

## Side Effects

- None.

## Errors

- Missing warning marker or session id fails.

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
	if !strings.Contains(body, "data-browser-agent-session-warning") {
		t.Fatalf("HTML missing data-browser-agent-session-warning; body=%s", truncate(body, 600))
	}
	if !strings.Contains(body, req.SessionID) {
		t.Fatalf("HTML missing session id %q; body=%s", req.SessionID, truncate(body, 500))
	}
	if !strings.Contains(body, "/v1/session") {
		t.Fatalf("HTML must reference /v1/session; body=%s", truncate(body, 500))
	}
}
```