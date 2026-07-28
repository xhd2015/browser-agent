## Expected

- Restore err nil.
- List empty.

## Side Effects

- None required (may or may not create sessions/).

## Errors

- Non-nil restore error fails this leaf.

## Exit Code

- 0.

```go
import (
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	assertRestoreOK(t, resp)
	if len(resp.ListIDs) != 0 {
		t.Fatalf("ListIDs=%v want empty", resp.ListIDs)
	}
	assertExitZero(t, resp)
}
```
