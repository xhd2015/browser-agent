## Expected

- `background.js` found under `Chrome-Ext-Browser-Agent`.
- At least one eager auto-attach path is present (register / create_tab /
  same-window navigate, or named eager helper).
- Those eager paths **guard** with `isCapturableTabURL` (or equivalent) so
  non-capturable URLs (`chrome://`, `chrome-extension://`, `devtools://`,
  `edge://`, `about:`) are not auto-attached.
- Capturable checks only inside job target picking do **not** satisfy this leaf.

## Side Effects

- None (read-only FS).

## Errors

- Blind eager attach on `chrome://` fails `chrome.debugger.attach` and may spam
  errors / leave partial attach state.

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
	if !hasEagerAttachSkipNonCapturable(text) {
		t.Fatalf("background eager-attach paths must skip non-capturable URLs (isCapturableTabURL on register/create/navigate attach); job pickTarget-only capturable checks are insufficient; text=%s",
			truncate(text, 900))
	}
}
```
