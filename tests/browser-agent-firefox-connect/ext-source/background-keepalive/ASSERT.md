## Expected

- `background.js` found under `Firefox-Ext-Browser-Agent`.
- Contains real **keepalive** and/or **reconnect** with **ping** (or alarm that reconnects
  dead WebSockets) — see `hasRealKeepalive`.
- Must **not** be only the P1 no-op alarm stub (`isNoOpOnlyAlarmStub` false when real path present).

## Side Effects

- None (read-only FS).

## Errors

- No-op-only 1-minute alarm leaves MV3/background WS dead after idle.

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
		t.Fatalf("Firefox background missing under ModuleRoot=%s; err=%q found=%v",
			req.ModuleRoot, resp.ErrText, resp.FoundPaths)
	}
	text := resp.CombinedText
	if isNoOpOnlyAlarmStub(text) {
		t.Fatalf("background keepalive must not be no-op-only alarm stub; text=%s", truncate(text, 700))
	}
	if !hasRealKeepalive(text) {
		t.Fatalf("background must implement real keepalive/reconnect/ping (not empty alarm); text=%s",
			truncate(text, 700))
	}
}
```
