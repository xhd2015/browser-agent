## Expected

- HTTP status **200**.
- Array length **2**.
- IDs sorted ascending: `sess-a`, `sess-b`.

## Side Effects

- Both sessions exist in registry.

## Errors

- Wrong order or missing id fails.

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
	if len(resp.SessionsListIDs) != 2 {
		t.Fatalf("sessions list len=%d want 2 ids=%v", len(resp.SessionsListIDs), resp.SessionsListIDs)
	}
	want := []string{"sess-a", "sess-b"}
	for i, id := range want {
		if resp.SessionsListIDs[i] != id {
			t.Fatalf("sessions[%d]=%q want %q full=%v", i, resp.SessionsListIDs[i], id, resp.SessionsListIDs)
		}
	}
}
```