## Expected

- Result ordered `["sess-a", "sess-c"]` (skip disconnected `sess-b`).

## Side Effects

- None.

## Errors

- Including disconnected id fails.

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
	assertIDsEqualOrdered(t, resp.ConnectedIDs, []string{"sess-a", "sess-c"})
	assertExitZero(t, resp)
}
```
