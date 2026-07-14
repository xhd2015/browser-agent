## Expected

- `background.js` found under `Chrome-Ext-Browser-Agent` (public, build, or src).
- Handles explicit **`tab_id`** from job payload (not only `pickTargetTabIdForSession` heuristics).
- Validates tab belongs to **`entry.windowId`** (session window scope).
- Does not target tabs from arbitrary windows when `tab_id` is set.

## Side Effects

- None (read-only FS).

## Errors

- Missing tab_id + windowId validation sends jobs to wrong tab/window.

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
	if !hasTabIDWindowValidation(text) {
		t.Fatalf("background must validate job tab_id against entry.windowId; text=%s",
			truncate(text, 900))
	}
}
```