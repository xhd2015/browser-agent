## Expected

Requirement scenario **#1** — control IP root:

- `ShouldCaptureURL("http://127.0.0.1:43759/")` returns **false**.

## Side Effects

- None (pure function).

## Errors

- Returning true would pollute capture with control traffic.

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
