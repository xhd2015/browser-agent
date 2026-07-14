## Expected

- `WSDisconnected` true (harness closed A socket).
- Inflight job on A: `HTTPJobOK` false; error mentions disconnect/connection/closed/lost.
- Session B snapshot: `ExtensionConnectedB` true after A disconnect.
- Follow-up job on B: `HTTPJobOKOnB` true.

## Side Effects

- A inflight job marked failed (not requeued).
- B socket and job queue unaffected.

## Errors

- A job ok=true after disconnect fails.
- B disconnected or B job failure fails "unaffected" contract.

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
		t.Fatal("WSDisconnected=false; harness should have closed session A socket")
	}
	if resp.HTTPJobOK {
		t.Fatal("HTTPJobOK=true after disconnect on A; v1 must fail inflight jobs")
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
		t.Fatalf("A job error should mention disconnect/connection loss; err=%q body=%s",
			resp.HTTPJobError, truncate(resp.BodyString, 400))
	}
	if !resp.ExtensionConnectedB {
		t.Fatalf("session B should stay connected after A disconnect; probe=%s", resp.SessionBProbeURL)
	}
	if !resp.HTTPJobOKOnB {
		t.Fatalf("B job should succeed after A disconnect; err=%q", resp.HTTPJobErrorOnB)
	}
}
```