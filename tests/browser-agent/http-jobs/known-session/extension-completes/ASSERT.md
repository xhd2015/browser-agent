## Expected

Requirement **C1**:

- HTTP status 200.
- Parsed job result `ok=true`.
- `HTTPJobError` empty (or absent).

## Side Effects

- Job may be audited under BaseDir (optional; not asserted).

## Errors

- Timeout or ok=false with extension auto-complete is a failure.

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
	if resp.HTTPJobError != "" {
		t.Fatalf("HTTPJobError should be empty on ok; got %q", resp.HTTPJobError)
	}
}
```
