## Expected

- `ValidateSessionID("a..b")` returns non-nil error.

## Side Effects

- None (pure).

## Errors

- Nil validate error fails this leaf.

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
	if resp.ValidateErr == nil {
		t.Fatalf("ValidateSessionID(%q) err=nil want non-nil", req.SessionID)
	}
	assertExitZero(t, resp)
}
```