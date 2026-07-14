## Expected

- HTTP status **404 Not Found**.

## Side Effects

- None.

## Errors

- 200 HTML for unknown session fails.

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