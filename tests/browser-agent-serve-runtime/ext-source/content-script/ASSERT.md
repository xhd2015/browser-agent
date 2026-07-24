## Expected

Requirement **E2**:

- contentScript found under Chrome-Ext-Browser-Agent.
- Text contains `__BROWSER_AGENT_EXT__`.
- Text contains `browser-agent` (feature or product marker).

## Side Effects

- None (read-only FS).

## Errors

- Missing page marker breaks product feature detection.

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
		t.Fatalf("contentScript missing under ModuleRoot=%s; err=%q found=%v",
			req.ModuleRoot, resp.ErrText, resp.FoundPaths)
	}
	text := resp.CombinedText
	if !strings.Contains(text, "__BROWSER_AGENT_EXT__") {
		t.Fatalf("contentScript must set __BROWSER_AGENT_EXT__; text=%s", truncate(text, 500))
	}
	if !strings.Contains(text, "browser-agent") {
		t.Fatalf("contentScript must mention browser-agent feature/product; text=%s", truncate(text, 500))
	}
	// Session attach: /go?session= page must register with the service worker.
	if !strings.Contains(text, "register") || !strings.Contains(text, "sendMessage") {
		t.Fatalf("contentScript must register session via chrome.runtime.sendMessage; text=%s", truncate(text, 600))
	}
	if !strings.Contains(text, "session_id") && !strings.Contains(text, "sessionId") {
		t.Fatalf("contentScript register payload must include session_id; text=%s", truncate(text, 600))
	}
}
```
