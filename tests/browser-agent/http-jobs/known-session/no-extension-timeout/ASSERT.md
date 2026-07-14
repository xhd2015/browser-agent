## Expected

Requirement **C2**:

- Response is not a silent success: `HTTPJobOK` is false **or** body/error mentions timeout.
- Prefer HTTP 200 with JobResult `{ok:false, error:…timeout…}`; also accept 408/504
  if error text still signals timeout.
- Must not be HTTP 404 (session is known).

## Side Effects

- Job status expired/failed server-side (optional).

## Errors

- ok=true without extension is a failure.
- Hanging until client hard cancel without timeout body is a failure.

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
	if resp.StatusCode == http.StatusNotFound {
		t.Fatalf("got 404 for known session; body=%s", truncate(resp.BodyString, 300))
	}
	if resp.HTTPJobOK {
		t.Fatal("HTTPJobOK=true without extension, want timeout failure")
	}
	msg := strings.ToLower(resp.HTTPJobError + " " + resp.BodyString)
	if !strings.Contains(msg, "timeout") {
		// Allow status-based timeout without body word if status is 408/504.
		if resp.StatusCode != http.StatusRequestTimeout && resp.StatusCode != http.StatusGatewayTimeout {
			t.Fatalf("expected timeout signal in body/error; status=%d err=%q body=%s",
				resp.StatusCode, resp.HTTPJobError, truncate(resp.BodyString, 400))
		}
	}
}
```
