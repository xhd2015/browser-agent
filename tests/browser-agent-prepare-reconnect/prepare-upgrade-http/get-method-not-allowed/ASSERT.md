## Expected

- HTTP status **405** Method Not Allowed.
- Must **not** be 200 (GET must not trigger broadcast).

## Side Effects

- No prepare_reconnect broadcast from GET.

## Errors

- 404 still fails (route must exist; only method is wrong). Prefer 405 over 404.

## Exit Code

- 0 (test harness).

```go
import (
	"net/http"
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	if resp == nil {
		t.Fatal("resp is nil")
	}
	if resp.StatusCode == http.StatusNotFound {
		t.Fatalf("status=404; route /v1/admin/prepare-upgrade missing — implement handler (GET should be 405)")
	}
	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Fatalf("status=%d want 405; body=%s", resp.StatusCode, resp.BodyString)
	}
	assertExitZero(t, resp)
}
```
