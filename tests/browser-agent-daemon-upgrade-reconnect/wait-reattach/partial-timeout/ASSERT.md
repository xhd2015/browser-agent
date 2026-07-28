## Expected

- Reattached contains `sess-a`.
- StillWaiting contains `sess-b`.
- Does not error (soft deadline).

## Side Effects

- None.

## Errors

- Wrong partition of ids fails.

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
	assertIDsEqual(t, resp.StillWaiting, []string{"sess-b"})
	assertExitZero(t, resp)
}
```
