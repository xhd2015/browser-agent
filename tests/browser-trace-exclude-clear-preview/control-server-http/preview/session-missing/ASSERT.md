## Expected

Requirement scenario **#6** — unknown session on preview:

- `GET /preview?session=does-not-exist` returns **HTTP 404**.
- Body indicates not found / session error (HTML or JSON).

## Side Effects

- Must not invent a preview page for an unknown session as 200 OK.

## Errors

- 200 HTML for a bogus session id is a failure.

## Exit Code

- Not asserted.

```go
import (
	"net/http"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("probe transport error: %v", err)
	}
	assertHTTPStatus(t, resp, http.StatusNotFound)
	assertNotFoundBody(t, resp.BodyString)
}
```
