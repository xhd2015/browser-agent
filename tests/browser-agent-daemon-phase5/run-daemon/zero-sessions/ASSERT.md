## Expected

- HTTP status **200** on `GET /v1/sessions`.
- Parsed session list length **0** (empty JSON array `[]`).

## Side Effects

- None beyond daemon lifecycle.

## Errors

- Non-empty session list fails.

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
	if len(resp.SessionsListIDs) != 0 {
		t.Fatalf("sessions len=%d want 0 ids=%v body=%s",
			len(resp.SessionsListIDs), resp.SessionsListIDs, truncate(resp.BodyString, 300))
	}
}
```