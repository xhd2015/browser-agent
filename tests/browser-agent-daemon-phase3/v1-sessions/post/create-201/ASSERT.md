## Expected

- HTTP status **201 Created**.
- JSON body includes `session_id` = `sess-create-201`.
- Body includes `session_url` containing `/go?session=sess-create-201`.

## Side Effects

- Session registered and artifacts written under `{BaseDir}/sessions/sess-create-201/`.

## Errors

- 200/409/400 fail this leaf.

## Exit Code

- Not asserted.

```go
import (
	"net/http"
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	assertHTTPStatus(t, resp, http.StatusCreated)
	assertJSONContentType(t, resp)
	if resp.CreatedSessionID != req.PostSessionID {
		t.Fatalf("session_id=%q want %q", resp.CreatedSessionID, req.PostSessionID)
	}
	if !strings.Contains(resp.CreatedSessionURL, "/go?session="+req.PostSessionID) {
		t.Fatalf("session_url=%q want /go?session=%s", resp.CreatedSessionURL, req.PostSessionID)
	}
}
```