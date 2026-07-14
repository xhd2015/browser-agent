## Expected

Requirement **#4** cell `(connected=true, supports=true)`:

- `ShouldExpandInstallPanel(true, true)` returns **false** (collapse).

## Side Effects

- None (pure function).

## Errors

- Returning true would leave the install panel open when the extension is healthy.

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
