## Expected

Requirement **P5** (default chrome):

- Source defines **`resolveInstallBrowser`**.
- Final / default return is **`"chrome"`** when no firefox signal matched.
- Must not default to firefox.

## Side Effects

- None (read-only FS).

## Errors

- Missing chrome default fails.

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

	if !hasResolveInstallBrowser(src) {
		t.Fatalf("resolveInstallBrowser must be defined; path=%v snippet=%s",
			resp.FoundPaths, truncate(src, 500))
	}

	body := extractNamedFuncBody(src, "resolveInstallBrowser")
	probe := body
	if probe == "" {
		probe = src
	}

	// Prefer trailing default return "chrome".
	hasChrome := strings.Contains(probe, `return "chrome"`) ||
		strings.Contains(probe, `return 'chrome'`) ||
		strings.Contains(probe, "return `chrome`") ||
		strings.Contains(probe, `return("chrome")`)
	if !hasChrome {
		// Fallback: function contains chrome return somewhere + return keyword near end.
		if strings.Contains(probe, `"chrome"`) && strings.Contains(probe, "return") {
			// Last return-ish: check last 200 chars prefer chrome over only firefox.
			tail := probe
			if len(tail) > 220 {
				tail = tail[len(tail)-220:]
			}
			if strings.Contains(tail, "chrome") {
				hasChrome = true
			}
		}
	}
	if !hasChrome {
		t.Fatalf("resolveInstallBrowser must default to \"chrome\"; path=%v body=%s",
			resp.FoundPaths, truncate(probe, 700))
	}
}
```
