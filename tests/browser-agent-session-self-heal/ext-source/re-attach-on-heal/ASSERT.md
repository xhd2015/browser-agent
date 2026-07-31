## Expected

- `background.js` found under `Chrome-Ext-Browser-Agent`.
- Heal / boot rediscover path reconnects sessions (register / connectSession).
- Heal body does **not** call `maybeEagerAttach` / `attachDebuggerForSession` /
  `listCapturableTabsInSessionWindow` attach loops (popup UX).

## Side Effects

- None (read-only FS).

## Errors

- Heal-time multi-attach is a regression for toolbar popup latency.

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
		t.Fatalf("shell background missing under ModuleRoot=%s; err=%q found=%v",
			req.ModuleRoot, resp.ErrText, resp.FoundPaths)
	}
	text := resp.CombinedText
	if !hasHealReconnectsWithoutDebuggerAttach(text) {
		t.Fatalf("background heal must reconnect WS without debugger attach (popup UX); text=%s",
			truncate(text, 900))
	}
}
```
