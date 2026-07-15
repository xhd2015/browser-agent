## Expected

Requirement **E2**:

- background.js found.
- Source mentions **`Target.`** prefix handling (e.g. `startsWith("Target.")`,
  `method.startsWith('Target.')`, `Target.createTarget`, or polyfill dispatch).
- Source mentions polyfill language **or** routes Target methods separately from
  generic debugger send (e.g. `polyfill`, `handleTarget`, `Target.` before
  sendCommand gate).
- Soft anti-pattern: pure fall-through of all CDP methods to sendCommand with
  **no** Target-specific branch is a failure.

## Side Effects

- None.

## Errors

- Missing Target intercept means Chrome -32000 Not allowed persists for Target.*.

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
		t.Fatalf("shell background missing; err=%q found=%v", resp.ErrText, resp.FoundPaths)
	}
	src := resp.CombinedText

	hasTargetPrefixGate := strings.Contains(src, `startsWith("Target.")`) ||
		strings.Contains(src, `startsWith('Target.')`) ||
		strings.Contains(src, "startsWith(\"Target.\")") ||
		strings.Contains(src, "\"Target.\"") ||
		strings.Contains(src, "'Target.'") ||
		strings.Contains(src, "Target.")

	hasPolyfillHint := strings.Contains(strings.ToLower(src), "polyfill") ||
		strings.Contains(src, "handleTarget") ||
		strings.Contains(src, "Target.createTarget") ||
		strings.Contains(src, "polyfillTarget") ||
		strings.Contains(src, "targetPolyfill")

	if !hasTargetPrefixGate {
		t.Fatalf("background must intercept Target.* methods; path=%v snippet=%s",
			resp.FoundPaths, truncate(src, 700))
	}
	if !hasPolyfillHint && !strings.Contains(src, "chrome.tabs") {
		t.Fatalf("background Target.* path must polyfill via chrome.tabs (or named polyfill); path=%v snippet=%s",
			resp.FoundPaths, truncate(src, 700))
	}
}
```
