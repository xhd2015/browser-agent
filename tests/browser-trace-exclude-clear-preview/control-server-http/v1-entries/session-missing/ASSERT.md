## Expected

Requirement scenario **#6** — unknown session on entries API:

- `GET /v1/entries?session=does-not-exist` returns **HTTP 404**.
- Body indicates not found / session error (JSON preferred).

## Side Effects

- Must not create a new session solely from a GET with unknown id.

## Errors

- 200 with empty entries would hide identity mistakes (failure).
- 500 without not-found semantics is a failure.

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
