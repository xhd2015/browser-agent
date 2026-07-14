## Expected

Requirement **G2**:

- Both session-page and popup entry files exist under `react/src/apps/`.
- FoundPaths length ≥ 2 (one per app).

## Side Effects

- None.

## Errors

- Only one app present fails multi-page product shell.

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
	if !resp.FileExists {
		t.Fatalf("missing session-page and/or popup app entries under react/src/apps/; found=%v ModuleRoot=%s",
			resp.FoundPaths, req.ModuleRoot)
	}
	if len(resp.FoundPaths) < 2 {
		t.Fatalf("want both session-page and popup entries; found=%v", resp.FoundPaths)
	}
	joined := strings.Join(resp.FoundPaths, "\n")
	if !strings.Contains(joined, "session-page") {
		t.Fatalf("missing session-page entry; found=%v", resp.FoundPaths)
	}
	if !strings.Contains(joined, "popup") {
		t.Fatalf("missing popup entry; found=%v", resp.FoundPaths)
	}
}
```
