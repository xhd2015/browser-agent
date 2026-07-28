## Expected

Requirement **R3** (Chrome path preserved):

- `InstallGuideline` source exists.
- Source still mentions **`chrome://extensions`** (Chrome default install).
- Soft prefer: Load unpacked / Developer mode still present for Chrome path.

## Side Effects

- None (read-only FS).

## Errors

- Removing Chrome install URL for default path fails product Chrome UX.

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
	if !hasChromeExtensionsURL(src) {
		t.Fatalf("InstallGuideline must preserve chrome://extensions for Chrome default path; path=%v snippet=%s",
			resp.FoundPaths, truncate(src, 700))
	}
	low := strings.ToLower(src)
	if !strings.Contains(low, "load unpacked") && !strings.Contains(low, "developer mode") {
		// Soft warning path: URL alone is the hard contract.
		t.Logf("note: Load unpacked / Developer mode not found alongside chrome://extensions; path=%v",
			resp.FoundPaths)
	}
}
```
