## Expected

- HTTP status **200**.
- `session_id` = `sess-known`.
- `extension.connected` false.
- `phase` indicates waiting when present (contains `wait` or `waiting_extension`).
- `hint` mentions `/go?session=sess-known` and keep-open / do-not-navigate guidance.

## Side Effects

- None.

## Errors

- 404 for known session fails.

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
	assertHTTPStatus(t, resp, http.StatusOK)
	if resp.SessionIDField != req.SessionID {
		t.Fatalf("session_id=%q want %q", resp.SessionIDField, req.SessionID)
	}
	if resp.ExtensionConnected {
		t.Fatal("extension.connected=true without hello")
	}
	if resp.Phase != "" {
		low := strings.ToLower(resp.Phase)
		if !strings.Contains(low, "wait") {
			t.Fatalf("phase=%q want waiting* before extension", resp.Phase)
		}
	}
	assertDisconnectedHint(t, resp.Hint, req.SessionID)
}
```