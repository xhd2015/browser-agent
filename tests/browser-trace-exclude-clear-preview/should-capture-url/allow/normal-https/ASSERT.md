## Expected

Requirement scenario **#2** — normal HTTPS:

- `ShouldCaptureURL("https://api.example.com/v1/users?limit=10")` returns **true**.

## Side Effects

- None (pure function).

## Errors

- False would drop legitimate app traffic from the capture buffer.

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
