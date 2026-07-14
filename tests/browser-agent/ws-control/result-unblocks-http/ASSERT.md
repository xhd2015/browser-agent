## Expected

Requirement **D3**:

- HTTP 200.
- `HTTPJobOK` true.
- `WSJobReceived` true (extension saw a job before answering).

## Side Effects

- None beyond short-lived job.

## Errors

- Timeout without result fails the leaf.

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
	if !resp.HTTPJobOK {
		t.Fatalf("HTTPJobOK=false; error=%q body=%s",
			resp.HTTPJobError, truncate(resp.BodyString, 400))
	}
	if !resp.WSJobReceived {
		t.Fatal("WSJobReceived=false; expected extension to observe job before result")
	}
}
```
