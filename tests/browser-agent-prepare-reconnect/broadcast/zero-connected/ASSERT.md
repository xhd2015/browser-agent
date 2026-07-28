## Expected

- `BroadcastPrepareReconnect` returns **empty** notified slice (nil or len 0).
- Run error is nil (skipping disconnected is not a failure).

## Side Effects

- No WS traffic.

## Errors

- Non-empty notified without connections fails.

## Exit Code

- 0.

```go
import (
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	if resp == nil {
		t.Fatal("resp is nil")
	}
	if len(resp.NotifiedIDs) != 0 {
		t.Fatalf("notified=%v want empty when zero connected", resp.NotifiedIDs)
	}
	assertExitZero(t, resp)
}
```
