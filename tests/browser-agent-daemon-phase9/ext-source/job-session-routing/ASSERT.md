## Expected

Requirement **P4**:

- `background.js` found under `Chrome-Ext-Browser-Agent`.
- Tab selection is scoped by **`session_id`** from job payload (e.g. `pickTargetTabId`
  accepts session id, or reads `payload.session_id` before tab query).
- Prefers **registered** `tabId` / `windowId` from sessions map when present.
- **Fallback** URL match for `/go?session=` (or `go?session=` + session id).

## Side Effects

- None (read-only FS).

## Errors

- Active-tab-only routing sends jobs to wrong session tab in multi-session daemon.

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
	low := strings.ToLower(text)
	hasSessionScopedPick := strings.Contains(low, "picktargettabid") ||
		(strings.Contains(low, "session_id") && strings.Contains(low, "tab"))
	if !hasSessionScopedPick {
		t.Fatalf("background must pick target tab scoped by session_id; text=%s", truncate(text, 600))
	}
	hasGoFallback := strings.Contains(low, "/go?session=") || strings.Contains(low, "go?session=")
	if !hasGoFallback {
		t.Fatalf("background must fallback to /go?session= URL match; text=%s", truncate(text, 600))
	}
	if !hasSessionsMap(text) {
		t.Fatalf("background must use sessions map for registered tab preference; text=%s", truncate(text, 600))
	}
}
```