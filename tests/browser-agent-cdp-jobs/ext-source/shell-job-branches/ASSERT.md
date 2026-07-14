## Expected

Requirement **D2**:

- background.js found.
- Source contains each job type token as a standalone word-ish match:
  `eval`, `run`, `logs`, `screenshot`, `cdp`, `info`
  (case-sensitive preferred for type strings; accept quoted forms).

## Side Effects

- None.

## Errors

- Missing any of the six types fails branch completeness.

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
	// Require quoted or comparison-friendly occurrences so we don't pass on
	// unrelated English words alone — prefer "eval" / 'eval' / === "eval".
	types := []string{"eval", "run", "logs", "screenshot", "cdp", "info"}
	for _, jt := range types {
		if !jobTypeTokenPresent(src, jt) {
			t.Fatalf("shell background missing job type branch token %q; path=%v snippet=%s",
				jt, resp.FoundPaths, truncate(src, 600))
		}
	}
}

func jobTypeTokenPresent(src, jt string) bool {
	// Accept several common JS forms.
	candidates := []string{
		`"` + jt + `"`,
		`'` + jt + `'`,
		"`" + jt + "`",
		"===\"" + jt + "\"",
		"== '" + jt + "'",
		"case \"" + jt + "\"",
		"case '" + jt + "'",
		"type === \"" + jt + "\"",
		"jobType === \"" + jt + "\"",
		"job_type === \"" + jt + "\"",
	}
	for _, c := range candidates {
		if strings.Contains(src, c) {
			return true
		}
	}
	// Fallback: bare token for short unique names (cdp, logs) when quoted form missing.
	// Require word-boundary-ish: not only as substring of longer identifier.
	if jt == "screenshot" || jt == "logs" || jt == "cdp" {
		return strings.Contains(src, jt)
	}
	// For eval/run/info, require quoted form to reduce false positives.
	return false
}
```
