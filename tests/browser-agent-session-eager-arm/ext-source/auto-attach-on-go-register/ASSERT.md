## Expected

- `background.js` found under `Chrome-Ext-Browser-Agent`.
- Session register path binds session and calls `connectSession`.
- Register path does **not** call `attachDebuggerForSession` / `maybeEagerAttach`
  (debugger attach on `/go` blocks toolbar popup UX).

## Side Effects

- None (read-only FS).

## Errors

- Register-time debugger attach on the control page is a product regression for popup.

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
	if !hasRegisterConnectsWithoutDebuggerAttach(text) {
		t.Fatalf("background must connectSession on /go register without debugger attach (popup UX); text=%s",
			truncate(text, 900))
	}
}
```
