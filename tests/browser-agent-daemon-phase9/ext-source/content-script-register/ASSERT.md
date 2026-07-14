## Expected

Requirement **P3**:

- `contentScript.js` found under `Chrome-Ext-Browser-Agent`.
- Reads session id from `/go?session=` page (URLSearchParams, `location.search`, or
  equivalent parsing `session` query param).
- Calls **`chrome.runtime.sendMessage`** with **`type:"register"`** (or `type: "register"`)
  and **`session_id`** plus tab/window identifiers.

## Side Effects

- None (read-only FS).

## Errors

- Marker-only content script without register breaks per-session WS attach.

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
		t.Fatalf("contentScript missing under ModuleRoot=%s; err=%q found=%v",
			req.ModuleRoot, resp.ErrText, resp.FoundPaths)
	}
	text := resp.CombinedText
	low := strings.ToLower(text)
	if !strings.Contains(low, "sendmessage") {
		t.Fatalf("contentScript must call chrome.runtime.sendMessage; text=%s", truncate(text, 500))
	}
	if !strings.Contains(low, "register") {
		t.Fatalf("contentScript must send register message type; text=%s", truncate(text, 500))
	}
	if !strings.Contains(low, "session_id") {
		t.Fatalf("contentScript must include session_id in register payload; text=%s", truncate(text, 500))
	}
	hasSessionFromURL := strings.Contains(low, "urlsearchparams") ||
		strings.Contains(low, "location.search") ||
		strings.Contains(low, "/go?session=") ||
		(strings.Contains(low, "session") && strings.Contains(low, "search"))
	if !hasSessionFromURL {
		t.Fatalf("contentScript must read session id from go page URL; text=%s", truncate(text, 500))
	}
}
```