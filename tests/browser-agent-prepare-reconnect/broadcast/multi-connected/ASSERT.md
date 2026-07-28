## Expected

- Notified set equals **`{sess-a, sess-b}`** (order-insensitive; implementer may sort).
- Both sockets receive `type=prepare_reconnect`.
- Both observe `delay_ms=1500`.

## Side Effects

- None beyond WS writes.

## Errors

- Missing either id in notified or either WS delivery fails.

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
	assertNotifiedEquals(t, resp.NotifiedIDs, []string{"sess-a", "sess-b"})
	for _, id := range []string{"sess-a", "sess-b"} {
		if !resp.WSReceivedBySession[id] {
			t.Fatalf("WS prepare_reconnect not received on %s; types=%v", id, resp.WSTypeBySession)
		}
		if resp.WSTypeBySession[id] != "prepare_reconnect" {
			t.Fatalf("%s WS type=%q want prepare_reconnect", id, resp.WSTypeBySession[id])
		}
		if resp.WSDelayMSBySession[id] != 1500 {
			t.Fatalf("%s delay_ms=%d want 1500; raw=%s",
				id, resp.WSDelayMSBySession[id], resp.WSPayloadRawBySession[id])
		}
	}
	assertExitZero(t, resp)
}
```
