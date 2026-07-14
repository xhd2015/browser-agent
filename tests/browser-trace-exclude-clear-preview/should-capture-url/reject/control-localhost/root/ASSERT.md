## Expected

Requirement scenario **#1** — localhost control root:

- `ShouldCaptureURL("http://localhost:43759/")` returns **false**.

## Side Effects

- None (pure function).

## Errors

- Returning true would miss localhost-shaped control traffic.

## Exit Code

- Not asserted.

```go
import (
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertCaptureResult(t, req, resp, err)
}
```
