## Expected

Requirement **G2** (Chrome inject path preserved):

- browseragent Go sources found.
- Combined inject/fallback surface still contains **`chrome://extensions`**.
- Soft: Load unpacked / data-browser-agent-install still present.

## Side Effects

- None (read-only FS).

## Errors

- Removing chrome://extensions from session-page inject fails Chrome UX.

## Exit Code

- Not asserted.

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	src := assertFilePresent(t, req, resp, "browseragent inject/fallback")
	if !hasChromeExtensionsURL(src) {
		t.Fatalf("inject/fallback must preserve chrome://extensions for Chrome default; paths=%v snippet=%s",
			resp.FoundPaths, truncate(src, 700))
	}
	if !strings.Contains(src, "data-browser-agent-install") &&
		!strings.Contains(strings.ToLower(src), "load unpacked") {
		t.Logf("note: install marker / Load unpacked not found with chrome://extensions; paths=%v",
			resp.FoundPaths)
	}
}
```
