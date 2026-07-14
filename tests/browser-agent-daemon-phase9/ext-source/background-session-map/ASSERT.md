## Expected

Requirement **P2**:

- `background.js` found under `Chrome-Ext-Browser-Agent`.
- Handles **`register`** message (via `chrome.runtime.onMessage` or equivalent).
- Maintains **`sessions`** map/object keyed by session id.
- Register flow references **`session_id`**, **`tabId`**, and **`windowId`**.

## Side Effects

- None (read-only FS).

## Errors

- Missing session map prevents routing jobs to the correct tab.

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
	text := resp.CombinedText
	if !hasRegisterHandler(text) {
		t.Fatalf("background must handle register message; text=%s", truncate(text, 600))
	}
	if !hasSessionsMap(text) {
		t.Fatalf("background must maintain sessions map; text=%s", truncate(text, 600))
	}
	assertContainsFold(t, text, "session_id", "tabid", "windowid")
}
```