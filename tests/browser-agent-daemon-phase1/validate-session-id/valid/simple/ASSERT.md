## Expected

- `ValidateSessionID("my-flow")` returns nil.
- `ValidateErr` is nil; `ValidateErrText` empty.

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
	if resp.ValidateErr != nil {
		t.Fatalf("ValidateSessionID(%q) err=%v want nil", req.SessionID, resp.ValidateErr)
	}
	assertExitZero(t, resp)
}
```