## Expected

- Notified equals **`{sess-connected}`** only — **not** `sess-orphan`.
- Connected socket receives `prepare_reconnect`.

## Side Effects

- Orphan session remains without WS traffic.

## Errors

- Including orphan in notified fails this leaf.

## Exit Code

- 0.

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	if resp == nil {
		t.Fatal("resp is nil")
	}
	assertNotifiedEquals(t, resp.NotifiedIDs, []string{"sess-connected"})
	for _, id := range resp.NotifiedIDs {
		if strings.Contains(id, "orphan") || id == "sess-orphan" {
			t.Fatalf("notified must not include orphan; got %v", resp.NotifiedIDs)
		}
	}
	if !resp.WSReceivedBySession["sess-connected"] {
		t.Fatalf("expected prepare_reconnect on sess-connected; types=%v", resp.WSTypeBySession)
	}
	if resp.WSTypeBySession["sess-connected"] != "prepare_reconnect" {
		t.Fatalf("type=%q want prepare_reconnect", resp.WSTypeBySession["sess-connected"])
	}
	assertExitZero(t, resp)
}
```
