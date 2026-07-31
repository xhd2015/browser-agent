## Expected

- `background.js` found under `Chrome-Ext-Browser-Agent`.
- `createTabInSession` (or `handleCreateTabJob` / shared create path) **auto-attaches**
  the newly created tab via `attachDebuggerForSession` (or named eager helper that
  does session-scoped attach).
- `chrome.tabs.create` alone does **not** satisfy this leaf.

## Side Effects

- None (read-only FS).

## Errors

- Without post-create attach, agents must fire a job before the new tab is in the
  multi-tab attach set — delays debugging and weakens policy B arming.

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
	if !hasAutoAttachOnCreateTab(text) {
		t.Fatalf("background must auto-attach new capturable tab after create_tab / createTabInSession (attachDebuggerForSession); tabs.create alone is insufficient; text=%s",
			truncate(text, 900))
	}
}
```
