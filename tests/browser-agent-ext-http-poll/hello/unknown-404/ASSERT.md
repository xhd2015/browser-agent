## Expected

- HTTP status **404**.
- Body is the control-plane session-not-found JSON (not a bare mux "404 page not found").
- Prefer `error`/`message` containing session-not-found wording.

## Side Effects

- No session created by hello.

## Errors

- 200, or 404 from missing route (`404 page not found`) fails.

## Exit Code

- 0 (assert on HTTP outcome; Run itself succeeds).

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
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status=%d want 404 for unknown session; body=%s",
			resp.StatusCode, truncate(resp.BodyString, 400))
	}
	// Reject httptest/mux default 404 so RED until the route is registered.
	if strings.Contains(strings.ToLower(resp.BodyString), "page not found") {
		t.Fatalf("got mux 404 page not found — POST /v1/ext/hello route not registered; body=%s",
			truncate(resp.BodyString, 200))
	}
	low := strings.ToLower(resp.BodyString)
	if !strings.Contains(low, "session") || !strings.Contains(low, "not found") {
		t.Fatalf("404 body should indicate session not found; body=%s", truncate(resp.BodyString, 400))
	}
	assertExitZero(t, resp)
}
```

