## Expected

Requirement **E1**:

- File exists at preferred path `react/src/lib/protocol/jobs.ts` (or `.js` / alt
  `react/src/protocol/jobs.*` accepted by harness).
- Combined source contains each job type string token:
  `info`, `eval`, `run`, `logs`, `screenshot`, `cdp`
  (quoted form preferred: `"info"` etc.).

## Side Effects

- None.

## Errors

- Missing module or missing any of the six types is a failure.

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
		t.Fatalf("protocol jobs module missing under ModuleRoot=%s; err=%q found=%v",
			req.ModuleRoot, resp.ErrText, resp.FoundPaths)
	}
	src := resp.CombinedText
	types := []string{"info", "eval", "run", "logs", "screenshot", "cdp"}
	for _, jt := range types {
		// Prefer quoted constants; also accept unquoted export names containing token.
		if strings.Contains(src, `"`+jt+`"`) || strings.Contains(src, `'`+jt+`'`) ||
			strings.Contains(src, "`"+jt+"`") {
			continue
		}
		// Fallback bare token (e.g. JOB_TYPE_EVAL = 'eval' already covered; bare last resort)
		if strings.Contains(strings.ToLower(src), jt) {
			continue
		}
		t.Fatalf("jobs module missing type token %q; path=%v snippet=%s",
			jt, resp.FoundPaths, truncate(src, 500))
	}
}
```
