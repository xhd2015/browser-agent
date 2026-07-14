## Expected

Requirement **D1**:

- background.js found under `Chrome-Ext-Browser-Agent`.
- Source contains:
  - `chrome.debugger` (attach/sendCommand/detach pattern preferred)
  - `Runtime.evaluate`
  - `Page.captureScreenshot`

## Side Effects

- None (read-only FS).

## Errors

- Stub-only background without CDP method names fails D1.

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
	src := resp.CombinedText
	needles := []string{
		"chrome.debugger",
		"Runtime.evaluate",
		"Page.captureScreenshot",
	}
	for _, n := range needles {
		if !strings.Contains(src, n) {
			t.Fatalf("shell background missing CDP token %q; path=%v snippet=%s",
				n, resp.FoundPaths, truncate(src, 500))
		}
	}
}
```
