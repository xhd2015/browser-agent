## Expected

Requirement **D4**:

- Content script file(s) found under Chrome-Ext-Capture-API (src or public).
- Combined text contains `__BROWSER_TRACE_EXT__`.
- Combined text contains `browser-trace`.

## Side Effects

- None.

## Errors

- Agent marker only would mis-advertise capture sessions.

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
		t.Fatalf("capture content script missing under Chrome-Ext-Capture-API; ModuleRoot=%s err=%q found=%v",
			req.ModuleRoot, resp.ErrText, resp.FoundPaths)
	}
	text := resp.CombinedText
	if !strings.Contains(text, "__BROWSER_TRACE_EXT__") {
		t.Fatalf("content script must define __BROWSER_TRACE_EXT__; paths=%v text=%s",
			resp.FoundPaths, truncate(text, 400))
	}
	if !strings.Contains(text, "browser-trace") {
		t.Fatalf("content script must mention product browser-trace; paths=%v text=%s",
			resp.FoundPaths, truncate(text, 400))
	}
}
```
