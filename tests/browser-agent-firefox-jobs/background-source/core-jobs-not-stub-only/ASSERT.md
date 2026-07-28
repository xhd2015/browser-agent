## Expected

Explicit anti-stub contract for Phase 2 core jobs:

1. `background.js` found.
2. **Must not** be Phase-1 job-stub-only (`isPhase1JobStubOnly` → false).
3. For each of **`info`**, **`eval`**, **`create_tab`**, **`screenshot`**:
   - job type branch token present
4. Real APIs present:
   - tab list path for info (`tabs.query` / list helper / `handleInfoJob`)
   - `tabs.create` for create_tab
   - `executeScript` for eval
   - `captureVisibleTab` for screenshot

## Side Effects

- None.

## Errors

- Phase-1 `not implemented: firefox job runner (phase 1 connect only)` as the only
  handleJob path for core types **fails**.

## Exit Code

- Not asserted.

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	src := assertBackgroundPresent(t, req, resp)

	if isPhase1JobStubOnly(src) {
		t.Fatalf("core jobs still Phase-1 stub-only (not implemented / missing real branches+APIs); path=%v snippet=%s",
			resp.FoundPaths, truncate(src, 800))
	}

	core := []string{"info", "eval", "create_tab", "screenshot"}
	for _, jt := range core {
		if !jobTypeTokenPresent(src, jt) {
			t.Fatalf("core job %q missing branch token; path=%v snippet=%s",
				jt, resp.FoundPaths, truncate(src, 600))
		}
	}

	// API gate (redundant with isPhase1JobStubOnly but clearer failure messages).
	if !hasTabsQueryOrList(src) && !strings.Contains(src, "handleInfoJob") {
		t.Fatalf("info must list tabs or define handleInfoJob; path=%v", resp.FoundPaths)
	}
	if !hasTabsCreate(src) {
		t.Fatalf("create_tab must call tabs.create; path=%v", resp.FoundPaths)
	}
	if !hasExecuteScript(src) {
		t.Fatalf("eval must use executeScript; path=%v", resp.FoundPaths)
	}
	if !hasCaptureVisibleTab(src) {
		t.Fatalf("screenshot must use captureVisibleTab; path=%v", resp.FoundPaths)
	}
}
```
