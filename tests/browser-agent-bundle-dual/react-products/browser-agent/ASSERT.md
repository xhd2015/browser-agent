## Expected

Requirement **E1**:

- File exists under `react/src/products/browser-agent.{ts,tsx,js}` (or fallback root).
- File content contains `43761`.
- File content contains `browser-agent`.

## Side Effects

- None.

## Errors

- Missing or wrong port breaks shared React parameterization.

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
		t.Fatalf("missing react/src/products/browser-agent.ts (or .tsx/.js) under ModuleRoot=%s",
			req.ModuleRoot)
	}
	text := resp.CombinedText
	if text == "" && len(resp.FoundPaths) > 0 && resp.FileContents != nil {
		text = resp.FileContents[resp.FoundPaths[0]]
	}
	if !strings.Contains(text, "43761") {
		t.Fatalf("product file must contain 43761; path=%s content=%s",
			resp.FoundPaths[0], truncate(text, 400))
	}
	if !strings.Contains(text, "browser-agent") {
		t.Fatalf("product file must contain browser-agent; path=%s content=%s",
			resp.FoundPaths[0], truncate(text, 400))
	}
}
```
