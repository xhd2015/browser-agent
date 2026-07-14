## Expected

- `background.js` reuses `chrome.debugger` attach when already attached to same `tabId`
  (`attachedTabs.has` or equivalent early return).
- **Detaches** when switching to a different `tab_id` between jobs
  (`chrome.debugger.detach` or wrapper).
- **Serializes** attach per session (lock/mutex/queue — no concurrent double-attach race).

## Side Effects

- None (read-only FS).

## Errors

- Missing detach-on-switch causes screenshot/eval on wrong tab or attach failures.

## Exit Code

- Not asserted.

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	if resp == nil {
		t.Fatal("resp is nil")
	}
	if !resp.FileExists || strings.TrimSpace(resp.CombinedText) == "" {
		t.Fatalf("shell background missing under ModuleRoot=%s; err=%q found=%v",
			req.ModuleRoot, resp.ErrText, resp.FoundPaths)
	}
	text := resp.CombinedText
	if !hasAttachReuseAndDetach(text) {
		t.Fatalf("background must reuse attach for same tab_id and detach on switch with serialized attach; text=%s",
			truncate(text, 900))
	}
}
```