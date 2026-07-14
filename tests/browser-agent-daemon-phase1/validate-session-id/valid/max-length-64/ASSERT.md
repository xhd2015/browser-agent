## Expected

- `ValidateSessionID` returns nil for a 64-character id.
- Id length is exactly 64.

## Side Effects

- None (pure).

## Errors

- Any non-nil validate error fails this leaf.

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
	if len(req.SessionID) != 64 {
		t.Fatalf("SessionID len=%d want 64", len(req.SessionID))
	}
	if resp.ValidateErr != nil {
		t.Fatalf("ValidateSessionID(64-char id) err=%v want nil", resp.ValidateErr)
	}
	assertExitZero(t, resp)
}
```