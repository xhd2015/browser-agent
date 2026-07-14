## Expected

- `POST /v1/shutdown` returns HTTP **202 Accepted**.

## Side Effects

- Shutdown may be in progress; harness cleanup cancels ctx if still running.

## Errors

- Non-202 status fails.

## Exit Code

- Not asserted.

```go
import (
	"net/http"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	assertHTTPStatus(t, resp, http.StatusAccepted)
}
```