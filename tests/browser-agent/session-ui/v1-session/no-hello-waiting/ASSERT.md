## Expected

Requirement **E1**:

- HTTP 200.
- `session_id` matches live session.
- `extension.connected` false.
- `supports_browser_agent` false.
- Phase indicates waiting (e.g. `waiting_extension`) when present.
- Hint non-empty preferred (waiting/install).

## Side Effects

- None.

## Errors

- Must not report connected without hello.

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
	wantID := resp.RealSessionID
	if wantID == "" {
		wantID = req.SessionID
	}
	if resp.SessionIDField != "" && resp.SessionIDField != wantID {
		t.Fatalf("session_id=%q, want %q", resp.SessionIDField, wantID)
	}
	if resp.ExtensionConnected {
		t.Fatal("extension.connected=true without hello")
	}
	if resp.SupportsBrowserAgent {
		t.Fatal("supports_browser_agent=true without hello")
	}
	if resp.Phase != "" && !strings.Contains(strings.ToLower(resp.Phase), "wait") {
		// Allow other waiting labels; only fail clearly active phases.
		if strings.Contains(strings.ToLower(resp.Phase), "connected") ||
			strings.Contains(strings.ToLower(resp.Phase), "recording") {
			t.Fatalf("phase=%q, want waiting* before hello", resp.Phase)
		}
	}
}
```
