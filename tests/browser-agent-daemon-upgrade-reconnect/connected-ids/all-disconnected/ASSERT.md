## Expected

- Result is empty (no connected ids).

## Side Effects

- None.

## Errors

- Any returned id fails.

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
	if len(resp.ConnectedIDs) != 0 {
		t.Fatalf("ConnectedIDs=%v want empty for all-disconnected", resp.ConnectedIDs)
	}
	assertExitZero(t, resp)
}
```
