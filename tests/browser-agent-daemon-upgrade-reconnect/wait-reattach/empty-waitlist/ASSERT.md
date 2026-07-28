## Expected

- `Reattached` and `StillWaiting` are empty.
- Does not hang until full timeout when waitList is empty (FetchCalls may be 0).

## Side Effects

- None.

## Errors

- Non-empty stillWaiting or reattached fails.

## Exit Code

- 0.

```go
import (
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	if resp == nil {
		t.Fatal("resp is nil")
	}
	if len(resp.Reattached) != 0 {
		t.Fatalf("Reattached=%v want empty", resp.Reattached)
	}
	if len(resp.StillWaiting) != 0 {
		t.Fatalf("StillWaiting=%v want empty", resp.StillWaiting)
	}
	// Empty waitList should not need polling.
	if resp.FetchCalls > 1 {
		t.Fatalf("FetchCalls=%d want 0 or at most 1 for empty waitList", resp.FetchCalls)
	}
	assertExitZero(t, resp)
}
```
