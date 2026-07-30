## Expected

- `background.js` found under `Chrome-Ext-Browser-Agent`.
- Attach path (`attachDebuggerForSession` / `withDebuggerForSession` or named gate helper)
  **checks** for an open session control tab before attaching.
- When no control tab remains, attach is **refused** with a clear error (throw/reject);
  no blind `chrome.debugger.attach`.

## Side Effects

- None (read-only FS).

## Errors

- Missing gate allows jobs to re-attach (or keep sticky attach) after the session page is gone.

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
	if !hasAttachGateRequiresSessionPage(text) {
		t.Fatalf("background must gate chrome.debugger.attach on open same-window /go?session= control tab and refuse when none; text=%s",
			truncate(text, 900))
	}
}
```
