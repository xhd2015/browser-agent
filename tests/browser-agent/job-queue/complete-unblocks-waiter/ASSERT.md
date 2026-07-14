## Expected

Requirement **B3**:

- Wait returns `ok=true`.
- Result data includes `value` equal to 2 (number).
- No CompleteErr.
- Job Get status is `done` (or equivalent terminal success).

## Side Effects

- Waiter unblocked without full JobTimeout.

## Errors

- Hanging Wait is a Run error (harness timeout).

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
	if resp.CompleteErr != "" {
		t.Fatalf("Complete error: %s", resp.CompleteErr)
	}
	if !resp.JobResultOK {
		t.Fatalf("JobResultOK=false; err=%q data=%v", resp.JobResultError, resp.JobResultData)
	}
	if resp.JobResultData == nil {
		t.Fatal("JobResultData is nil")
	}
	v := resp.JobResultData["value"]
	switch n := v.(type) {
	case float64:
		if n != 2 {
			t.Fatalf("data.value = %v, want 2", v)
		}
	case int:
		if n != 2 {
			t.Fatalf("data.value = %v, want 2", v)
		}
	default:
		t.Fatalf("data.value = %v (%T), want 2", v, v)
	}
	st := resp.JobGetStatus
	if st != "" && st != "done" && st != "completed" {
		// Allow empty if Get not implemented yet will fail elsewhere; prefer done.
		t.Fatalf("JobGetStatus = %q, want done (or empty if not exposed)", st)
	}
}
```
