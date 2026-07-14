## Expected

- `ValidateSessionID` returns non-nil error for 65-character id.
- Id length is exactly 65.

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
	if len(req.SessionID) != 65 {
		t.Fatalf("SessionID len=%d want 65", len(req.SessionID))
	}
	if resp.ValidateErr == nil {
		t.Fatal("ValidateSessionID(65-char id) err=nil want non-nil")
	}
	assertExitZero(t, resp)
}
```