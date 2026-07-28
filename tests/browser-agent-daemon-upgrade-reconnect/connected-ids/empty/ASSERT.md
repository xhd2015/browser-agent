## Expected

- `ConnectedIDsFromSnapshots(nil)` returns an empty list (len 0).
- No panic.

## Side Effects

- None (pure function).

## Errors

- Non-empty result fails this leaf.

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
		t.Fatalf("ConnectedIDs=%v want empty", resp.ConnectedIDs)
	}
	assertExitZero(t, resp)
}
```
