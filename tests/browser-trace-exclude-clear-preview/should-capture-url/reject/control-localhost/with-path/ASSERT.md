## Expected

Requirement scenario **#1** — localhost control path:

- `ShouldCaptureURL("http://localhost:43759/preview?session=x")` returns **false**.

## Side Effects

- None (pure function).

## Errors

- True would capture the live preview document request.

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
