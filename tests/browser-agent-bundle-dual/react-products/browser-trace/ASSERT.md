## Expected

Requirement **E2**:

- File exists under `react/src/products/browser-trace.{ts,tsx,js}` (or fallback root).
- File content contains `43759`.
- File content contains `browser-trace`.

## Side Effects

- None.

## Errors

- Missing dual product file collapses coexistence design.

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
	if !resp.FileExists || len(resp.FoundPaths) == 0 {
		t.Fatalf("missing react/src/products/browser-trace.ts (or .tsx/.js) under ModuleRoot=%s",
			req.ModuleRoot)
	}
	text := resp.CombinedText
	if text == "" && len(resp.FoundPaths) > 0 && resp.FileContents != nil {
		text = resp.FileContents[resp.FoundPaths[0]]
	}
	if !strings.Contains(text, "43759") {
		t.Fatalf("product file must contain 43759; path=%s content=%s",
			resp.FoundPaths[0], truncate(text, 400))
	}
	if !strings.Contains(text, "browser-trace") {
		t.Fatalf("product file must contain browser-trace; path=%s content=%s",
			resp.FoundPaths[0], truncate(text, 400))
	}
}
```
