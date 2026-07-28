## Expected

- HTTP status **404**.
- Body is session-not-found JSON (not mux `404 page not found`).

## Side Effects

- None.

## Errors

- 200, or bare mux 404 (route missing) fails.

## Exit Code

- 0.

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
	if strings.Contains(strings.ToLower(resp.BodyString), "page not found") {
		t.Fatalf("got mux 404 page not found — POST /v1/ext/poll route not registered; body=%s",
			truncate(resp.BodyString, 200))
	}
	low := strings.ToLower(resp.BodyString)
	if !strings.Contains(low, "session") || !strings.Contains(low, "not found") {
		t.Fatalf("404 body should indicate session not found; body=%s", truncate(resp.BodyString, 400))
	}
	assertExitZero(t, resp)
}
```

