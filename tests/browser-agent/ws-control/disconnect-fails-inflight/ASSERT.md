## Expected

Requirement **D4** + DSN disconnect policy v1:

- `WSDisconnected` true (harness closed the socket).
- Job outcome is not success: `HTTPJobOK` false.
- Error/body mentions disconnect, connection, closed, or lost (case-insensitive).
- Must not hang until the full multi-second job timeout without failing earlier
  (soft: as long as ok=false with disconnect signal).

## Side Effects

- Inflight job marked failed (not requeued for a second attempt).

## Errors

- ok=true after disconnect is a failure.
- Treating disconnect as silent timeout-only without connection signal is weak;
  prefer disconnect wording.

## Exit Code

- Not asserted.

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	if resp == nil {
		t.Fatal("resp is nil")
	}
	if !resp.WSDisconnected {
		t.Fatal("WSDisconnected=false; harness should have closed the socket")
	}
	if resp.HTTPJobOK {
		t.Fatal("HTTPJobOK=true after disconnect; v1 must fail inflight jobs")
	}
	msg := strings.ToLower(resp.HTTPJobError + " " + resp.BodyString)
	ok := strings.Contains(msg, "disconnect") ||
		strings.Contains(msg, "connection") ||
		strings.Contains(msg, "connect") ||
		strings.Contains(msg, "closed") ||
		strings.Contains(msg, "lost") ||
		strings.Contains(msg, "websocket") ||
		strings.Contains(msg, "ws ")
	if !ok {
		t.Fatalf("error should mention disconnect/connection loss; err=%q body=%s",
			resp.HTTPJobError, truncate(resp.BodyString, 400))
	}
}
```
