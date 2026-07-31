## Expected

- `background.js` found under `Chrome-Ext-Browser-Agent`.
- `createTabInSession` scopes `chrome.tabs.create` to **`entry.windowId`**.
- At least one eager auto-attach path exists (create_tab and/or same-window
  navigate) and is **session-window scoped** (`entry.windowId` / tab.windowId
  match) — no proactive auto-attach for tabs in other windows.
- Window-scoped create alone (without eager attach) does **not** satisfy P2.

## Side Effects

- None (read-only FS).

## Errors

- Global auto-attach across all windows would attach unrelated user windows and
  expand the debugger banner / attach set beyond the session window.

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
	if !hasEagerAttachScopedToSessionWindow(text) {
		t.Fatalf("background must scope create + eager attach to entry.windowId (no other-window proactive auto-attach); create-only without eager attach is insufficient; text=%s",
			truncate(text, 900))
	}
}
```
