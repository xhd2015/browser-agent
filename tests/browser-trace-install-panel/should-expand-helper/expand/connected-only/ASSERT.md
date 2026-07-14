## Expected

Requirement **#4** cell `(connected=true, supports=false)`:

- `ShouldExpandInstallPanel(true, false)` returns **true**.

## Side Effects

- None (pure function).

## Errors

- Returning false would collapse update guidance for an unsupported extension.

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
