## Expected

- HTTP status **200**.
- At least one event with `type` **`prepare_reconnect`** (`HasPrepareReconnect`).

## Side Effects

- Event is drained (or at least visible) for this poll.

## Errors

- Empty events or WS-only broadcast (no poll event) fails.

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
		t.Fatalf("status=%d want 200; body=%s",
			resp.StatusCode, truncate(resp.BodyString, 500))
	}
	if !resp.HasPrepareReconnect {
		t.Fatalf("expected events to include type=prepare_reconnect for HTTP-connected session; "+
			"EventCount=%d body=%s (BroadcastPrepareReconnect must enqueue poll events)",
			resp.EventCount, truncate(resp.BodyString, 500))
	}
	assertJSONContentType(t, resp.ContentType)
	assertExitZero(t, resp)
}
```
