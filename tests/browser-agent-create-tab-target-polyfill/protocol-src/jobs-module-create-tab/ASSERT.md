## Expected

Requirement **P1**:

- File exists at preferred path `react/src/lib/protocol/jobs.ts` (or `.js` / alt).
- Combined source contains job type string token **`create_tab`** (quoted preferred).
- Preferred also: `JOB_TYPE_CREATE_TAB` and/or membership in `KNOWN_JOB_TYPES`.

## Side Effects

- None.

## Errors

- Missing module or missing `create_tab` token is a failure.

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
	hasQuoted := strings.Contains(src, `"create_tab"`) ||
		strings.Contains(src, `'create_tab'`) ||
		strings.Contains(src, "`create_tab`")
	hasConst := strings.Contains(src, "JOB_TYPE_CREATE_TAB") ||
		strings.Contains(src, "create_tab")
	if !hasQuoted && !hasConst {
		t.Fatalf("jobs module missing create_tab token; path=%v snippet=%s",
			resp.FoundPaths, truncate(src, 500))
	}
	// Prefer quoted form for protocol fidelity; allow bare if constant name present.
	if !hasQuoted && !strings.Contains(src, "JOB_TYPE_CREATE_TAB") {
		t.Fatalf("jobs module should quote \"create_tab\" or define JOB_TYPE_CREATE_TAB; path=%v snippet=%s",
			resp.FoundPaths, truncate(src, 500))
	}
}
```
