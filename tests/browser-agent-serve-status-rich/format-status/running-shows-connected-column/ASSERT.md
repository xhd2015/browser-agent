## Expected

- `FormatDaemonStatus` returns **nil** error.
- Sessions table header includes **`Connected`** column (with `Session ID`, `Phase`).
- Fixture session row includes **`no`** (extension not connected yet).

## Side Effects

- Formatter only.

## Errors

- Missing Connected header/cell fails.

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
	if resp.FormatErr != "" {
		t.Fatalf("FormatDaemonStatus error: %s", resp.FormatErr)
	}
	if resp.Formatted == "" {
		t.Fatal("formatted output is empty")
	}
	if !resp.Status.Running {
		t.Fatal("expected running status for format fixture")
	}
	low := strings.ToLower(resp.Formatted)
	if !strings.Contains(low, "connected") {
		t.Fatalf("formatted output missing Connected column; got:\n%s", truncate(resp.Formatted, 800))
	}
	if !strings.Contains(low, "session id") || !strings.Contains(low, "phase") {
		t.Fatalf("formatted output missing session table headers; got:\n%s", truncate(resp.Formatted, 800))
	}
	assertContainsFold(t, resp.Formatted, req.SessionID, "no")
}
```