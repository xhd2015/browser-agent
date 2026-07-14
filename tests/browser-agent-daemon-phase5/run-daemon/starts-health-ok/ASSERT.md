## Expected

- HTTP status **200** on `GET /v1/health`.

## Side Effects

- RunDaemon started and shut down cleanly via harness cleanup.

## Errors

- Non-200 health fails.

## Exit Code

- Not asserted.

```go
import (
	"net/http"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	assertHTTPStatus(t, resp, http.StatusOK)
}
```