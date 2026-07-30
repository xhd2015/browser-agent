## Expected

- `background.js` found under `Chrome-Ext-Browser-Agent` (public, build, or src).
- Session leave / `unregisterSession` path **detaches** debugger held by that session
  (`detachDebugger` / `chrome.debugger.detach`, or a named session-detach helper).
- Leave detach is **not** satisfied only by tab-switch detach inside
  `attachDebuggerForSession`.

## Side Effects

- None (read-only FS).

## Errors

- Missing leave-detach leaves the Chrome “started debugging this page” banner after
  the last session control tab is gone.

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
	if !hasDetachOnSessionLeave(text) {
		t.Fatalf("background must detach session debugger on last session-page leave / unregisterSession (not only WS teardown or tab-switch detach); text=%s",
			truncate(text, 900))
	}
}
```
