## Expected

Requirement **#4** cell `(connected=false, supports=true)`:

- `ShouldExpandInstallPanel(false, true)` returns **true**.

## Side Effects

- None (pure function).

## Errors

- Returning false would hide install steps for a not-connected session.

## Exit Code

- Not asserted.

```go
import (
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertExpandResult(t, req, resp, err)
}
```
