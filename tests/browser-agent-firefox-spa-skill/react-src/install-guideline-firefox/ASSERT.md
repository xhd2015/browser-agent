## Expected

Requirement **R1** (InstallGuideline Firefox path):

- `InstallGuideline` source exists under `react/src/ui/` (or components).
- Source documents Firefox temporary add-on install:
  1. **`about:debugging`** (full `about:debugging#/runtime/this-firefox` OK)
  2. Temporary add-on wording (**Load Temporary Add-on** / temporary add-on)
  3. Canonical path segment **`browser-agent-firefox`**
- Accept dual-path component (Chrome + Firefox branches) as long as Firefox
  markers are present in this file.

## Side Effects

- None (read-only FS).

## Errors

- Missing any of the three markers fails Firefox SPA install UX.

## Exit Code

- Not asserted.

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	src := assertFilePresent(t, req, resp, "InstallGuideline")
	if !strings.Contains(src, "InstallGuideline") && !strings.Contains(strings.ToLower(src), "install-guideline") {
		// Soft: filename may encode component; body should still be the guideline.
		t.Logf("note: InstallGuideline identifier not found in body; path=%v", resp.FoundPaths)
	}

	if !hasAboutDebugging(src) {
		t.Fatalf("InstallGuideline missing about:debugging for Firefox install; path=%v snippet=%s",
			resp.FoundPaths, truncate(src, 700))
	}
	if !hasTemporaryAddon(src) {
		t.Fatalf("InstallGuideline missing temporary add-on / Load Temporary Add-on; path=%v snippet=%s",
			resp.FoundPaths, truncate(src, 700))
	}
	if !hasFirefoxExtPath(src) {
		t.Fatalf("InstallGuideline missing browser-agent-firefox path segment; path=%v snippet=%s",
			resp.FoundPaths, truncate(src, 700))
	}
}
```
