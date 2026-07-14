## Expected

Requirement **B2**:

- Two dequeued types in order: `eval` then `info`.
- Two non-empty ids; first id ≠ second id.

## Side Effects

- None.

## Errors

- Reverse order is a FIFO failure.

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
	if len(resp.DequeuedTypes) != 2 {
		t.Fatalf("DequeuedTypes = %v, want [eval info]", resp.DequeuedTypes)
	}
	if resp.DequeuedTypes[0] != "eval" || resp.DequeuedTypes[1] != "info" {
		t.Fatalf("FIFO order types = %v, want [eval info]", resp.DequeuedTypes)
	}
	if len(resp.DequeuedIDs) != 2 || resp.DequeuedIDs[0] == "" || resp.DequeuedIDs[1] == "" {
		t.Fatalf("DequeuedIDs = %v, want two non-empty", resp.DequeuedIDs)
	}
	if resp.DequeuedIDs[0] == resp.DequeuedIDs[1] {
		t.Fatalf("job ids must be unique; both %q", resp.DequeuedIDs[0])
	}
}
```
