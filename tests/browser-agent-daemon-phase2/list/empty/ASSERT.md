## Expected

- `List()` on fresh registry returns empty slice (length 0).

## Side Effects

- None.

## Errors

- Non-empty list fails this leaf.

## Exit Code

- 0.

```go
import (
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	if resp == nil {
		t.Fatal("resp is nil")
	}
	if len(resp.ListSessionIDs) != 0 {
		t.Fatalf("List len=%d want 0 ids=%v", len(resp.ListSessionIDs), resp.ListSessionIDs)
	}
	assertExitZero(t, resp)
}
```