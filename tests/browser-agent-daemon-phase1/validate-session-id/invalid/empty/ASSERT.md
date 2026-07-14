## Expected

- `ValidateSessionID("")` returns non-nil error.
- Error message is non-empty (descriptive).

## Side Effects

- None (pure).

## Errors

- Nil validate error fails this leaf.

## Exit Code

- 0 (API error returned in Response, not Run transport error).

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
		t.Fatal("ValidateSessionID(\"\") err=nil want non-nil")
	}
	if resp.ValidateErrText == "" {
		t.Fatal("expected descriptive error text")
	}
	assertExitZero(t, resp)
}
```