## Expected

Requirement **D3**:

- Content script file(s) found under Chrome-Ext-Browser-Agent.
- Combined text contains `__BROWSER_AGENT_EXT__`.
- Combined text contains `browser-agent` (product id string).

## Side Effects

- None.

## Errors

- Using `__BROWSER_TRACE_EXT__` only would break agent page detection.

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
		t.Fatalf("agent content script missing under Chrome-Ext-Browser-Agent; ModuleRoot=%s err=%q found=%v",
			req.ModuleRoot, resp.ErrText, resp.FoundPaths)
	}
	text := resp.CombinedText
	if !strings.Contains(text, "__BROWSER_AGENT_EXT__") {
		t.Fatalf("content script must define __BROWSER_AGENT_EXT__; paths=%v text=%s",
			resp.FoundPaths, truncate(text, 400))
	}
	if !strings.Contains(text, "browser-agent") {
		t.Fatalf("content script must mention product browser-agent; paths=%v text=%s",
			resp.FoundPaths, truncate(text, 400))
	}
}
```
