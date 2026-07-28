## Expected

- Result is exactly `["sess-aaa", "sess-bbb"]` in that order.

## Side Effects

- None.

## Errors

- Wrong set or wrong order fails.

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
	assertIDsEqualOrdered(t, resp.ConnectedIDs, []string{"sess-aaa", "sess-bbb"})
	assertExitZero(t, resp)
}
```
