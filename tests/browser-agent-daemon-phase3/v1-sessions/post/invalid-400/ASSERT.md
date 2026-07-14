## Expected

- HTTP status **400 Bad Request**.
- Body indicates invalid session id (soft: status mandatory).

## Side Effects

- No session registered for invalid id.

## Errors

- 409 or 201 fails.

## Exit Code

- Not asserted.

```go
import (
	"net/http"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	assertHTTPStatus(t, resp, http.StatusBadRequest)
}
```