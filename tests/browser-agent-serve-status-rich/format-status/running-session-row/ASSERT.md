## Expected

- `FormatDaemonStatus` returns **nil** error.
- Output contains created **`Session ID`** and session **`Phase`** for fixture session.
- `Sessions (N)` count is at least 1.

## Side Effects

- Formatter only.

## Errors

- Missing session row fails.

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
	if len(resp.Status.Sessions) < 1 {
		t.Fatalf("expected at least one session in status; got %d", len(resp.Status.Sessions))
	}
	phase := strings.TrimSpace(resp.Status.Sessions[0].Phase)
	if phase == "" {
		t.Fatal("session phase empty in status snapshot")
	}
	assertContainsFold(t, resp.Formatted, req.SessionID, phase, "sessions (")
}
```