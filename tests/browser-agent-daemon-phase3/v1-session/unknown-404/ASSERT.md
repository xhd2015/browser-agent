## Expected

- HTTP status **404 Not Found**.
- Body indicates unknown / not found (soft: status mandatory).

## Side Effects

- None.

## Errors

- 200 with snapshot fails.

## Exit Code

- Not asserted.

```go
import (
	"net/http"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	assertHTTPStatus(t, resp, http.StatusNotFound)
}
```