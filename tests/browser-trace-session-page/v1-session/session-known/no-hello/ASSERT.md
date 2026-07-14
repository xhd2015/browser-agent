# Expected

Requirement scenario **#1** — no hello yet:

- HTTP 200, JSON content type.
- `session_id` equals the live session id (`req.SessionSuffix` / `resp.RealSessionID`).
- `phase` is `waiting_extension`.
- `extension.connected` is **false**.
- `extension.supports_browser_trace` is **false**.
- `extension.version` is empty (or absent/empty string).
- `recording.active` is false; `entry_count` is 0.
- `hint` is non-empty and mentions waiting for the extension and/or install/enable.
- `ready.deadline_ms` > 0; remaining/elapsed are non-negative.

## Side Effects

- None beyond harness teardown of the short-lived session.

## Errors

- Must not report connected=true without a prior hello.
- Must not invent supports_browser_trace=true without capability inputs.

## Exit Code

- Not asserted.

```go
import (
	"net/http"
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("probe transport error: %v", err)
	}
	assertHTTPStatus(t, resp, http.StatusOK)
	assertJSONContentType(t, resp)

	wantID := resp.RealSessionID
	if wantID == "" {
		wantID = req.SessionSuffix
	}
	assertSessionIDMatch(t, resp, wantID)
	assertPhase(t, resp, "waiting_extension")
	assertExtensionConnected(t, resp, false)
	assertSupportsBrowserTrace(t, resp, false)
	if resp.ExtensionVersion != "" {
		t.Fatalf("extension.version = %q, want empty before hello", resp.ExtensionVersion)
	}
	assertRecording(t, resp, false, 0)
	assertReadyCountdownPresent(t, resp)
	assertHintNonEmpty(t, resp)

	h := strings.ToLower(resp.Hint)
	ok := strings.Contains(h, "wait") ||
		strings.Contains(h, "install") ||
		strings.Contains(h, "extension")
	if !ok {
		t.Fatalf("hint should mention waiting/install/extension; got %q", resp.Hint)
	}
}
```
