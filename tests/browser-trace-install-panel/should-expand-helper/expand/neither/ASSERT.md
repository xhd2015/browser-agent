## Expected

Requirement **#4** cell `(connected=false, supports=false)`:

- `ShouldExpandInstallPanel(false, false)` returns **true**.

## Side Effects

- None (pure function).

## Errors

- Returning false would collapse the panel while the extension is not working.

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
