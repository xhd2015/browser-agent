## Expected

- HTTP status **404 Not Found**.
- Body indicates unknown session (soft: status mandatory).

## Side Effects

- No job executed for live sessions.

## Errors

- 200 with ok=false is insufficient — unknown session must be 404.

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