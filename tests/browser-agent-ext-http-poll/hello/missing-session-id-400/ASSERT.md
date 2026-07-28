## Expected

- HTTP status **400**.

## Side Effects

- Session remains disconnected.

## Errors

- 200 attach without session_id fails the test.

## Exit Code

- 0.

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
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status=%d want 400 for missing session_id; body=%s",
			resp.StatusCode, truncate(resp.BodyString, 400))
	}
	assertExitZero(t, resp)
}
```
