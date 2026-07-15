## Expected

Requirement **T3** (React `SessionPageApp`):

- Source file exists (`react/src/ui/SessionPageApp.tsx` preferred).
- Source assigns **`document.title`** using the session id and suffix
  **` - Browser Agent`** (exact spaces around hyphen).
- Accept any of:
  - `document.title = \`${sid} - Browser Agent\``
  - `document.title = sid + " - Browser Agent"`
  - equivalent with `sessionId` / `snap?.session_id` when that is the known sid
- **Empty / missing sid (scenario 4)**: must not set a broken sole title
  `" - Browser Agent"`. Prefer guard `if (!sid) return` / only assign inside
  a branch where sid is truthy. Hard-fail if source clearly does
  `document.title = " - Browser Agent"` or
  `document.title = \` - Browser Agent\`` with no sid interpolation.

## Side Effects

- None (read-only FS).

## Errors

- Missing `document.title`, missing suffix, or empty-sid broken title fails.

## Exit Code

- Not asserted.

```go
import (
	"regexp"
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	if resp == nil {
		t.Fatal("resp is nil")
	}
	if !resp.FileExists || strings.TrimSpace(resp.CombinedText) == "" {
		t.Fatalf("SessionPageApp source missing; err=%q found=%v ModuleRoot=%s",
			resp.ErrText, resp.FoundPaths, req.ModuleRoot)
	}
	src := resp.CombinedText

	if !strings.Contains(src, "document.title") {
		t.Fatalf("SessionPageApp must set document.title; path=%v snippet=%s",
			resp.FoundPaths, truncate(src, 500))
	}

	// Must include product suffix token.
	if !strings.Contains(src, " - Browser Agent") && !strings.Contains(src, TitleSuffix) {
		t.Fatalf("document.title path must use suffix %q; path=%v snippet=%s",
			TitleSuffix, resp.FoundPaths, truncate(src, 600))
	}

	// Must tie title to session id somehow (sid / sessionId / session_id).
	hasSidRef := strings.Contains(src, "sid") ||
		strings.Contains(src, "sessionId") ||
		strings.Contains(src, "session_id")
	if !hasSidRef {
		t.Fatalf("document.title assignment should reference session id (sid/sessionId); path=%v",
			resp.FoundPaths)
	}

	// Plausible assignment patterns (template or concat).
	patterns := []*regexp.Regexp{
		regexp.MustCompile(`document\.title\s*=\s*` + "`[^`]*\\$\\{[^}]+\\}[^`]*Browser Agent`"),
		regexp.MustCompile(`document\.title\s*=\s*[^;\n]+Browser Agent`),
		regexp.MustCompile(`document\.title\s*=\s*[^;\n]+` + regexp.QuoteMeta(TitleSuffix)),
	}
	matched := false
	for _, re := range patterns {
		if re.MatchString(src) {
			matched = true
			break
		}
	}
	if !matched {
		// Fallback: document.title and suffix appear in same function-ish window.
		idx := strings.Index(src, "document.title")
		if idx >= 0 {
			end := idx + 200
			if end > len(src) {
				end = len(src)
			}
			window := src[idx:end]
			if strings.Contains(window, "Browser Agent") {
				matched = true
			}
		}
	}
	if !matched {
		t.Fatalf("could not find document.title assignment with Browser Agent suffix; path=%v snippet=%s",
			resp.FoundPaths, truncate(src, 700))
	}

	// Empty-sid guard: forbid broken sole title with no interpolation.
	broken := []string{
		`document.title = " - Browser Agent"`,
		`document.title=" - Browser Agent"`,
		"document.title = ` - Browser Agent`",
		"document.title=` - Browser Agent`",
		`document.title = ' - Browser Agent'`,
	}
	for _, b := range broken {
		if strings.Contains(src, b) {
			t.Fatalf("must not set broken empty-sid title; found %q in path=%v", b, resp.FoundPaths)
		}
	}
}
```
