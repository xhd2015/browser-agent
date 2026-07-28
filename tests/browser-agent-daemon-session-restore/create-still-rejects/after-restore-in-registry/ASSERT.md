## Expected

- Restore succeeds; Get(`sess-dup`) true.
- Create(`sess-dup`) returns `errors.Is(err, ErrSessionExists)`.

## Side Effects

- Session remains registered; dir remains.

## Errors

- Create succeeds or wrong error type fails this leaf.

## Exit Code

- 0.

```go
import (
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	assertRestoreOK(t, resp)
	assertGetOK(t, resp, "sess-dup", true)
	assertErrSessionExists(t, resp.CreateErr)
	if !resp.CreateErrIsSessionExists {
		t.Fatal("CreateErrIsSessionExists=false want true")
	}
	assertExitZero(t, resp)
}
```
