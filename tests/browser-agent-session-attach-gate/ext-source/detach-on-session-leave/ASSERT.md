## Expected

- `background.js` found under `Chrome-Ext-Browser-Agent` (public, build, or src).
- Session leave / `unregisterSession` path **detaches every tab** held in that session’s
  attach set (`attachedTabIds` / equivalent — iterate + detach, not only one sticky id).
- Leave detach is **not** satisfied only by tab-switch detach inside
  `attachDebuggerForSession`.
- Pure sticky `attachedTabId` one-shot detach does **not** pass (policy B).

## Side Effects

- None (read-only FS).

## Errors

- Missing leave-detach leaves the Chrome “started debugging this page” banner after
  the last session control tab is gone.
- Detaching only one sticky id while peers remain attached leaves orphaned debuggees.

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
	if !hasDetachAllOnSessionLeave(text) {
		t.Fatalf("background must detach ALL tabs in session attach set on last session-page leave / unregisterSession (not only sticky single attachedTabId); text=%s",
			truncate(text, 900))
	}
}
```
