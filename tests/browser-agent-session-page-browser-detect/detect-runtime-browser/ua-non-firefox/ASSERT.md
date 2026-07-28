## Expected

Requirement **D2** (`detectRuntimeBrowser` non-Firefox default):

- Source defines **`detectRuntimeBrowser`**.
- Default / else path returns **`"chrome"`** when UA is not Firefox
  (Chrome UA, empty string, missing navigator).
- Must not hard-code only `"firefox"` without a chrome fallback.

## Side Effects

- None (read-only FS).

## Errors

- Missing helper or missing chrome default return fails.

## Exit Code

- Not asserted.

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	src := assertSourcePresent(t, req, resp, "SessionPageApp/helpers")

	if !hasDetectRuntimeBrowser(src) {
		t.Fatalf("detectRuntimeBrowser must be defined; path=%v snippet=%s",
			resp.FoundPaths, truncate(src, 500))
	}

	body := extractNamedFuncBody(src, "detectRuntimeBrowser")
	probe := body
	if probe == "" {
		probe = src
	}

	hasChromeReturn := strings.Contains(probe, `"chrome"`) ||
		strings.Contains(probe, `'chrome'`) ||
		strings.Contains(probe, "`chrome`")
	if !hasChromeReturn {
		t.Fatalf("detectRuntimeBrowser must return \"chrome\" for non-Firefox UA; path=%v body=%s",
			resp.FoundPaths, truncate(probe, 600))
	}

	// Sanity: still a dual-outcome helper (not chrome-only stub without firefox).
	if !strings.Contains(probe, "firefox") && !strings.Contains(probe, "Firefox") {
		t.Fatalf("detectRuntimeBrowser should also handle Firefox path; path=%v body=%s",
			resp.FoundPaths, truncate(probe, 600))
	}
}
```
