## Expected

Requirement **W2** (wire resolve result into InstallGuideline):

- SessionPageApp (or combined SPA) **calls** `resolveInstallBrowser(...)`.
- Result is stored (e.g. `installBrowser`) and passed to **`InstallGuideline`**
  as `browser={...}` / `browser=installBrowser` / equivalent.
- Soft: still imports/uses `InstallGuideline`.

## Side Effects

- None (read-only FS).

## Errors

- resolveInstallBrowser unused, or InstallGuideline without browser binding fails.

## Exit Code

- Not asserted.

```go
import (
	"regexp"
	"strings"
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	src := assertSourcePresent(t, req, resp, "SessionPageApp")

	if !strings.Contains(src, "resolveInstallBrowser") {
		t.Fatalf("SessionPageApp must call resolveInstallBrowser; path=%v snippet=%s",
			resp.FoundPaths, truncate(src, 500))
	}
	if !strings.Contains(src, "InstallGuideline") {
		t.Fatalf("SessionPageApp must render InstallGuideline; path=%v snippet=%s",
			resp.FoundPaths, truncate(src, 500))
	}

	// Call site: resolveInstallBrowser(...)
	callRe := regexp.MustCompile(`resolveInstallBrowser\s*\(`)
	if !callRe.MatchString(src) {
		t.Fatalf("expected resolveInstallBrowser(...) call site; path=%v snippet=%s",
			resp.FoundPaths, truncate(src, 600))
	}

	// browser={installBrowser} or browser={...} near InstallGuideline.
	// Accept common patterns.
	patterns := []*regexp.Regexp{
		regexp.MustCompile(`(?s)InstallGuideline[\s\S]{0,400}browser\s*=\s*\{[^}]*installBrowser`),
		regexp.MustCompile(`(?s)browser\s*=\s*\{[^}]*installBrowser[\s\S]{0,400}InstallGuideline`),
		regexp.MustCompile(`browser=\{installBrowser\}`),
		regexp.MustCompile(`browser=\{[^}]*resolveInstallBrowser`),
	}
	ok := false
	for _, re := range patterns {
		if re.MatchString(src) {
			ok = true
			break
		}
	}
	if !ok {
		// Fallback: both InstallGuideline and browser= with installBrowser in file.
		if strings.Contains(src, "InstallGuideline") &&
			strings.Contains(src, "installBrowser") &&
			(strings.Contains(src, "browser={") || strings.Contains(src, "browser = {") || strings.Contains(src, "browser={installBrowser}")) {
			ok = true
		}
	}
	if !ok {
		t.Fatalf("InstallGuideline must receive browser={installBrowser} (or resolve result); path=%v snippet=%s",
			resp.FoundPaths, truncate(src, 900))
	}
}
```
