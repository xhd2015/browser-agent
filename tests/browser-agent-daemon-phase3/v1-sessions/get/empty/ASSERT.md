## Expected

- HTTP status **200**.
- JSON array length **0**.

## Side Effects

- None.

## Errors

- Non-array body or non-empty list fails.

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
	assertJSONContentType(t, resp)
	if len(resp.SessionsListIDs) != 0 {
		t.Fatalf("sessions list len=%d want 0 ids=%v", len(resp.SessionsListIDs), resp.SessionsListIDs)
	}
}
```