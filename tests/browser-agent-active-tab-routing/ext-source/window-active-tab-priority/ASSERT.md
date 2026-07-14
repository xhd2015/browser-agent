## Expected

- `background.js` found under `Chrome-Ext-Browser-Agent` (public, build, or src).
- Defines **`pickTargetTabIdForSession`** (or equivalent session-scoped picker).
- Queries **active tab in session window**: `active: true` combined with
  `windowId: entry.windowId` (or `entry.windowId` in the query).
- **Session-page fallback** still present: `entry.tabId` and `/go?session=` URL scan.
- Routing is **not** limited to global last-focused window only.

## Side Effects

- None (read-only FS).

## Errors

- Missing active+windowId query sends jobs to session control page when user tab is active.

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
	if !hasActiveTabInWindowQuery(text) {
		t.Fatalf("background must query active tab in entry.windowId via pickTargetTabIdForSession; text=%s",
			truncate(text, 800))
	}
	if !hasSessionPageFallback(text) {
		t.Fatalf("background must retain session-page fallback (entry.tabId and /go?session=); text=%s",
			truncate(text, 800))
	}
	low := strings.ToLower(text)
	if strings.Contains(low, "lastfocusedwindow") && !strings.Contains(low, "entry.windowid") {
		t.Fatalf("background must scope active tab to session window (entry.windowId), not only lastFocusedWindow; text=%s",
			truncate(text, 800))
	}
}
```