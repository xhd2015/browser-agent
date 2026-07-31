## Expected

- `background.js` reuses `chrome.debugger` attach when already attached to same `tabId`
  (`attachedTabs.has` or equivalent early return).
- **Serializes** attach per session (lock/mutex/queue — no concurrent double-attach race).
- Does **not** require detach-on-switch of a different previously attached tab
  (policy B multi-attach; peers kept).

## Side Effects

- None (read-only FS).

## Errors

- Missing reuse/serialize causes screenshot/eval attach failures or races.

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
	if !hasAttachReuseAndSerialize(text) {
		t.Fatalf("background must reuse attach for same tab_id and serialize attach (no switch-detach demand); text=%s",
			truncate(text, 900))
	}
}
```
