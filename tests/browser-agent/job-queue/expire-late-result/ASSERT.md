## Expected

Requirement **B6**:

- Wait outcome is not ok (JobResultOK false) and/or error mentions timeout.
- `JobGetStatus` is `expired` (preferred) or another non-success terminal (`failed`).
- Status must **not** be `done` / `completed` after late Complete.
- LateCompleteErr may be empty (ignored) or non-empty (rejected); either is fine.

## Side Effects

- No panic on late Complete (Run would fail).

## Errors

- Accepting late ok as success is a failure.

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
	if resp.JobResultOK {
		t.Fatal("JobResultOK=true after expire path, want false")
	}
	st := strings.ToLower(resp.JobGetStatus)
	if st == "done" || st == "completed" {
		t.Fatalf("JobGetStatus=%q after late complete; must not be successful done", resp.JobGetStatus)
	}
	if st != "" && st != "expired" && st != "failed" {
		// Prefer expired; allow failed; empty means Get missing — soft fail with message.
		t.Fatalf("JobGetStatus=%q, want expired (preferred) or failed", resp.JobGetStatus)
	}
	if st == "" {
		// Require at least timeout signal on the wait path if status not exposed.
		msg := strings.ToLower(resp.JobResultError)
		if !strings.Contains(msg, "timeout") && !strings.Contains(msg, "expir") {
			t.Fatalf("empty JobGetStatus and error %q lacks timeout/expire signal", resp.JobResultError)
		}
	}
}
```
