## Expected

- HTTP status **200**.
- `ok` true.
- `notified` is **empty** (nil or `[]`).

## Side Effects

- None.

## Errors

- Non-200 or non-empty notified without connections fails.

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
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status=%d want 200; body=%s", resp.StatusCode, resp.BodyString)
	}
	if !resp.HTTPOK {
		t.Fatalf("ok field false or missing; body=%s", resp.BodyString)
	}
	if len(resp.HTTPNotifiedIDs) != 0 {
		t.Fatalf("notified=%v want empty; body=%s", resp.HTTPNotifiedIDs, resp.BodyString)
	}
	assertExitZero(t, resp)
}
```
