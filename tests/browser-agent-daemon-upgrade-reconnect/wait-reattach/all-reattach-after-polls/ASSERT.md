## Expected

- Reattached equals waitList; StillWaiting empty.
- FetchCalls ≥ 3 (did not stop on first disconnected poll).

## Side Effects

- None.

## Errors

- Early exit or timeout without full reattach fails.

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
	assertIDsEqual(t, resp.Reattached, []string{"sess-a", "sess-b"})
	if len(resp.StillWaiting) != 0 {
		t.Fatalf("StillWaiting=%v want empty", resp.StillWaiting)
	}
	if resp.FetchCalls < 3 {
		t.Fatalf("FetchCalls=%d want >=3 (delayed reattach)", resp.FetchCalls)
	}
	assertExitZero(t, resp)
}
```
