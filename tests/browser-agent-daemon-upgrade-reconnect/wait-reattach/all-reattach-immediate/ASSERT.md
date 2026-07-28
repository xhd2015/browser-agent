## Expected

- Reattached set equals waitList `{sess-a, sess-b}`.
- StillWaiting empty.
- Completes with few fetch calls (typically 1).

## Side Effects

- None.

## Errors

- Missing reattached id or non-empty stillWaiting fails.

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
	if resp.FetchCalls < 1 {
		t.Fatal("expected at least one fetch call")
	}
	assertExitZero(t, resp)
}
```
