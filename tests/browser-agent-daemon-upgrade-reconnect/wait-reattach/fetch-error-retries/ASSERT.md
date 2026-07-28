## Expected

- Reattached contains `sess-a`; StillWaiting empty.
- FetchCalls ≥ 3 (retried past errors).
- No transport/Run error (fetch errors are internal retries).

## Side Effects

- None.

## Errors

- Treating first fetch error as hard failure fails this leaf.

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
	assertIDsEqual(t, resp.Reattached, []string{"sess-a"})
	if len(resp.StillWaiting) != 0 {
		t.Fatalf("StillWaiting=%v want empty after successful retry", resp.StillWaiting)
	}
	if resp.FetchCalls < 3 {
		t.Fatalf("FetchCalls=%d want >=3 (error retries)", resp.FetchCalls)
	}
	assertExitZero(t, resp)
}
```
