## Expected

- `FormatDaemonStatus` returns **nil** error.
- Output contains case-insensitive markers: **`status`**, **`sessions`**.
- Output includes table header tokens (e.g. **`pid`**, **`addr`**, **`session`**).
- When running, output mentions created session id.

## Side Effects

- Formatter writes only to provided writer (no meta changes).

## Errors

- Format error or missing markers fails.

## Exit Code

- Not asserted.

```go
import (
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
	assertContainsFold(t, resp.Formatted, "status", "sessions")
	assertContainsFold(t, resp.Formatted, "pid", "addr", "session")
	if !resp.Status.Running {
		t.Fatal("expected running status for format fixture")
	}
	assertContainsFold(t, resp.Formatted, req.SessionID)
}
```