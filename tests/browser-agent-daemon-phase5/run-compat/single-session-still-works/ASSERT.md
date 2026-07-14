## Expected

- HTTP status **200** on `GET /v1/session?session=sess-p5-compat`.
- `session_id` matches `sess-p5-compat`.
- `extension.connected` false before hello.
- Probe URL includes `?session=sess-p5-compat`.

## Side Effects

- Run started and shut down cleanly via harness cleanup.

## Errors

- 400/404 for live Run session fails.

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
	if resp.SessionIDField != req.SessionID {
		t.Fatalf("session_id=%q want %q", resp.SessionIDField, req.SessionID)
	}
	if resp.ExtensionConnected {
		t.Fatal("extension.connected=true without hello")
	}
	if !strings.Contains(resp.ProbeURL, "session="+req.SessionID) {
		t.Fatalf("ProbeURL=%q want session query for %q", resp.ProbeURL, req.SessionID)
	}
}
```