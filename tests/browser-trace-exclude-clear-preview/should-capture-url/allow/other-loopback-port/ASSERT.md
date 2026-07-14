## Expected

Requirement scenario **#2** (boundary) — other loopback port:

- `ShouldCaptureURL("http://127.0.0.1:8080/api/local")` returns **true**.
- Exclusion is host **and** port `43759`, not all of loopback.

## Side Effects

- None (pure function).

## Errors

- False would hide local app traffic when developers hit 127.0.0.1:8080.

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
