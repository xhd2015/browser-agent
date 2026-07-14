## Expected

- HTTP status **400 Bad Request**.
- Body indicates missing/empty session parameter (soft: status mandatory).

## Side Effects

- None.

## Errors

- 200 defaulting to arbitrary session fails in multi-session mode.

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