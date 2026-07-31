## Expected

- `background.js` found under `Chrome-Ext-Browser-Agent`.
- `tabs.onUpdated` and/or `tabs.onCreated` neighborhood (or a named eager helper
  wired from those listeners) **auto-attaches** capturable tabs in the armed
  session window via `attachDebuggerForSession` (or equivalent).
- Control-tab `/go` register alone (`maybeRegisterGoTab` without attach) does
  **not** satisfy this leaf.

## Side Effects

- None (read-only FS).

## Errors

- Without same-window navigate attach, manually opened tabs in the session window
  stay outside the attach set until a job targets them.

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
	if !hasAutoAttachOnSameWindowNavigate(text) {
		t.Fatalf("background must auto-attach capturable tabs on same-window navigate/update while session armed (tabs.onUpdated/onCreated → attachDebuggerForSession); maybeRegisterGoTab alone is insufficient; text=%s",
			truncate(text, 900))
	}
}
```
