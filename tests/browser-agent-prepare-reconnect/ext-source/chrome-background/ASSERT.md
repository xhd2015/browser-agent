## Expected

Chrome `background.js` found and contains markers for:

1. **`prepare_reconnect`** (or `prepareReconnect`) message type handling.
2. **`delay_ms`** (or `delayMs`) read from payload.
3. **Schedule close** after delay (`setTimeout`/`alarms` + `close`).
4. **Aggressive reconnect** after close (reconnect attempt reset / force reconnect /
   retry base / prepare_reconnect path).

## Side Effects

- None (read-only FS).

## Errors

- Missing file or any required marker fails (Phase 2 extension contract).

## Exit Code

- Not asserted.

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
	if !resp.FileExists || strings.TrimSpace(resp.CombinedText) == "" {
		t.Fatalf("Chrome background.js missing under ModuleRoot=%s; err=%q found=%v",
			req.ModuleRoot, resp.ErrText, resp.FoundPaths)
	}
	text := resp.CombinedText
	if !hasPrepareReconnectTypeMarker(text) {
		t.Fatalf("Chrome background must handle prepare_reconnect; path=%v snippet=%s",
			resp.FoundPaths, truncate(text, 700))
	}
	if !hasDelayMsMarker(text) {
		t.Fatalf("Chrome background must read delay_ms; path=%v snippet=%s",
			resp.FoundPaths, truncate(text, 700))
	}
	if !hasScheduleCloseMarker(text) {
		t.Fatalf("Chrome background must schedule close after delay_ms (setTimeout/alarms + close); path=%v snippet=%s",
			resp.FoundPaths, truncate(text, 700))
	}
	if !hasAggressiveReconnectMarker(text) {
		t.Fatalf("Chrome background must aggressive-reconnect after prepare_reconnect close; path=%v snippet=%s",
			resp.FoundPaths, truncate(text, 700))
	}
}
```
