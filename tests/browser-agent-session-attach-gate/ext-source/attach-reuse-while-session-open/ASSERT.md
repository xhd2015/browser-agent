## Expected

- `background.js` reuses `chrome.debugger` attach when already attached to the same
  `tabId` (`attachedTabs.has` / set `.has` / equivalent early return).
- **Serializes** attach per session (`attachLock` or equivalent).
- Does **not** require switch-detach of a different previously attached tab.

## Side Effects

- None (read-only FS).

## Errors

- Losing reuse forces detach-after-every-job (out of scope) or double-attach races.

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
	if !hasSameTabReuseAndLock(text) {
		t.Fatalf("background must reuse attach for same tab while session open and serialize attach (attachLock); text=%s",
			truncate(text, 900))
	}
}
```
