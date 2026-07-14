## Expected

Requirement **G3**:

- InstallGuideline component file exists under `react/src/ui/` (or
  `react/src/components/`).
- Path basename contains `InstallGuideline`.

## Side Effects

- None.

## Errors

- Missing component breaks session-page install UX sharing.

## Exit Code

- Not asserted.

```go
import (
	"path/filepath"
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	if resp == nil {
		t.Fatal("resp is nil")
	}
	if !resp.FileExists || len(resp.FoundPaths) == 0 {
		t.Fatalf("missing InstallGuideline component under react/src/ui/; ModuleRoot=%s",
			req.ModuleRoot)
	}
	base := filepath.Base(resp.FoundPaths[0])
	if !strings.Contains(base, "InstallGuideline") {
		t.Fatalf("unexpected path basename %q; want InstallGuideline*", base)
	}
}
```
