## Expected

Requirement scenario **#1** — control IP with path/query:

- `ShouldCaptureURL("http://127.0.0.1:43759/v1/entries?session=abc")` returns **false**.

## Side Effects

- None (pure function).

## Errors

- True would capture the agent’s own entries push requests.

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
