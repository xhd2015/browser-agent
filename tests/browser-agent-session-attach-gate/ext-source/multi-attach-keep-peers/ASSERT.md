## Expected

- `background.js` found under `Chrome-Ext-Browser-Agent`.
- Per-session attach state is a **set** of tab IDs (`attachedTabIds` / equivalent),
  not only a singular sticky `attachedTabId`.
- `attachDebuggerForSession` records the new `tabId` into that set.
- Attach path does **not** switch-detach the previous session-attached tab solely
  because a different `tabId` is being attached (policy A obsolete).

## Side Effects

- None (read-only FS).

## Errors

- Sticky switch-detach tears down peer CDP attaches mid-session when jobs target
  different tabs.

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
	if !hasMultiAttachKeepPeers(text) {
		t.Fatalf("background must multi-attach (set of tab ids) and keep peer tabs attached (no switch-detach); text=%s",
			truncate(text, 900))
	}
}
```
