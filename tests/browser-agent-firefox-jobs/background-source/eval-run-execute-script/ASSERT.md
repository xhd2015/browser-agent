## Expected

- `background.js` found.
- Job type tokens **`eval`** and **`run`** present.
- Script injection API: **`executeScript`** via `scripting.executeScript` **or**
  `tabs.executeScript` (browser/chrome prefix OK).

## Side Effects

- None.

## Errors

- Debugger-only Runtime.evaluate without executeScript is **not** Phase 2 GREEN.
- Phase-1 stub without executeScript fails.

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
	if !jobTypeTokenPresent(src, "eval") {
		t.Fatalf("background missing eval job branch token; path=%v snippet=%s",
			resp.FoundPaths, truncate(src, 600))
	}
	if !jobTypeTokenPresent(src, "run") {
		t.Fatalf("background missing run job branch token; path=%v snippet=%s",
			resp.FoundPaths, truncate(src, 600))
	}
	if !hasExecuteScript(src) {
		t.Fatalf("eval/run must use scripting.executeScript or tabs.executeScript; path=%v snippet=%s",
			resp.FoundPaths, truncate(src, 600))
	}
	// Soft: prefer not requiring chrome.debugger for Phase 2; if only debugger path, still fail above.
	_ = strings.Contains(src, "chrome.debugger")
}
```
