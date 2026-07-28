## Expected

- `BroadcastPrepareReconnect` returns notified ids containing exactly **`sess-a`**.
- Fake extension on `sess-a` receives WS envelope `type=prepare_reconnect`.
- Payload `delay_ms` is **1000**.

## Side Effects

- Other sessions (none) unchanged.

## Errors

- Empty notified or missing WS delivery fails.

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
	assertNotifiedEquals(t, resp.NotifiedIDs, []string{"sess-a"})
	if !resp.WSReceivedBySession["sess-a"] {
		t.Fatalf("WS prepare_reconnect not received on sess-a; types=%v raw=%v",
			resp.WSTypeBySession, resp.WSPayloadRawBySession)
	}
	if resp.WSTypeBySession["sess-a"] != "prepare_reconnect" {
		t.Fatalf("WS type=%q want prepare_reconnect", resp.WSTypeBySession["sess-a"])
	}
	if resp.WSDelayMSBySession["sess-a"] != 1000 {
		t.Fatalf("WS delay_ms=%d want 1000; raw=%s",
			resp.WSDelayMSBySession["sess-a"], resp.WSPayloadRawBySession["sess-a"])
	}
	assertExitZero(t, resp)
}
```
