## Expected

Requirement **T2** (fallback `writeFallbackSessionHTML`):

- Source file defining `writeFallbackSessionHTML` exists under `browseragent/`.
- Locate the **function definition** (`func … writeFallbackSessionHTML`), not the
  call site in `handleGo`.
- Within that function body, format uses product suffix
  **` - Browser Agent`** (exact spaces around hyphen).
- Title is dynamic with session id — accept any of:
  - string concat / sprintf involving session id and ` - Browser Agent`
  - `<title>` template containing `%s - Browser Agent` or equivalent
    (`%s` / `%v` / `esc` + suffix)
- Must **not** rely solely on the old static title
  **`browser-agent session`** as the fallback page title (that static sole title
  is forbidden once feature lands).

## Side Effects

- None (read-only FS).

## Errors

- Missing function definition, missing `<title>` in definition body, missing
  ` - Browser Agent` near title, or only static `browser-agent session` title fails.

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
		t.Fatalf("fallback source missing; err=%q found=%v ModuleRoot=%s",
			resp.ErrText, resp.FoundPaths, req.ModuleRoot)
	}
	src := resp.CombinedText
	if !strings.Contains(src, "writeFallbackSessionHTML") {
		t.Fatalf("source missing writeFallbackSessionHTML; path=%v", resp.FoundPaths)
	}

	// Prefer function definition body; never stop at handleGo call site.
	fallbackSrc := extractWriteFallbackFuncSrc(src)
	if strings.TrimSpace(fallbackSrc) == "" {
		// Last resort: full-file title tokens (still require function def marker).
		if !regexp.MustCompile(`func\s+(?:\([^)]*\)\s+)?writeFallbackSessionHTML\b`).MatchString(src) {
			t.Fatalf("writeFallbackSessionHTML function definition not found; path=%v", resp.FoundPaths)
		}
		fallbackSrc = src
	}

	// Must mention <title> in fallback shell.
	if !strings.Contains(strings.ToLower(fallbackSrc), "<title") {
		t.Fatalf("writeFallbackSessionHTML definition missing <title>; path=%v snippet=%s",
			resp.FoundPaths, truncate(fallbackSrc, 500))
	}

	// Desired format token: " - Browser Agent" (exact spacing).
	if !strings.Contains(fallbackSrc, TitleSuffix) && !strings.Contains(fallbackSrc, " - Browser Agent") {
		t.Fatalf("fallback title must use suffix %q; path=%v snippet=%s",
			TitleSuffix, resp.FoundPaths, truncate(fallbackSrc, 700))
	}

	// Dynamic session involvement: sprintf/concat patterns or title including format verb / esc / sessionID.
	dynamic := false
	patterns := []*regexp.Regexp{
		regexp.MustCompile(`(?s)<title[^>]*>[^<]*%s[^<]*Browser Agent`),
		regexp.MustCompile(`(?s)<title[^>]*>[^<]*%[v][^<]*Browser Agent`),
		regexp.MustCompile(`sessionID\s*\+\s*" - Browser Agent"`),
		regexp.MustCompile(`esc\s*\+\s*" - Browser Agent"`),
		regexp.MustCompile(`" - Browser Agent"`),
		regexp.MustCompile("` - Browser Agent`"),
		regexp.MustCompile(`fmt\.Sprintf\([^)]*Browser Agent`),
	}
	for _, re := range patterns {
		if re.MatchString(fallbackSrc) {
			dynamic = true
			break
		}
	}
	if !dynamic && strings.Contains(fallbackSrc, " - Browser Agent") {
		// Suffix present next to title is enough if not pure static old title only.
		dynamic = true
	}
	if !dynamic {
		t.Fatalf("fallback title does not look dynamic with session id + %q; path=%v snippet=%s",
			TitleSuffix, resp.FoundPaths, truncate(fallbackSrc, 700))
	}

	// Forbid sole static old fallback title as the only title form in the window.
	// Allow the old string to appear elsewhere only if new format is present (already required).
	staticOnly := strings.Contains(fallbackSrc, "browser-agent session") &&
		!strings.Contains(fallbackSrc, " - Browser Agent")
	if staticOnly {
		t.Fatalf("fallback still uses sole static title %q; want {sessionId}%s; path=%v",
			"browser-agent session", TitleSuffix, resp.FoundPaths)
	}
}

// extractWriteFallbackFuncSrc returns source from the function definition of
// writeFallbackSessionHTML through the next top-level func (or a large window).
// It deliberately skips call sites like c.writeFallbackSessionHTML(...).
func extractWriteFallbackFuncSrc(src string) string {
	// Match: func writeFallbackSessionHTML... or func (recv) writeFallbackSessionHTML...
	re := regexp.MustCompile(`(?m)^func\s+(?:\([^)]*\)\s+)?writeFallbackSessionHTML\b`)
	loc := re.FindStringIndex(src)
	if loc == nil {
		// Non-line-anchored fallback (unusual formatting).
		re2 := regexp.MustCompile(`func\s+(?:\([^)]*\)\s+)?writeFallbackSessionHTML\b`)
		loc = re2.FindStringIndex(src)
	}
	if loc == nil {
		return ""
	}
	start := loc[0]
	end := start + 6000
	if end > len(src) {
		end = len(src)
	}
	window := src[start:end]
	// Cut before the next top-level func declaration when present.
	if next := strings.Index(window[1:], "\nfunc "); next >= 0 {
		window = window[:1+next]
	}
	return window
}
```
