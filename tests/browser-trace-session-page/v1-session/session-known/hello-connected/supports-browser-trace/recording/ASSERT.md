# Expected

Requirement scenario **#4** — after status recording:

- HTTP 200 JSON.
- `phase` is `recording`.
- `recording.active` is **true**.
- `recording.entry_count` equals the posted count (`7`).
- `recording.window_id` equals the posted window id (`42`) when exposed.
- `extension.connected` true; `supports_browser_trace` true; version still echoed.

## Side Effects

- None asserted (HAR lifecycle is covered by `./tests/browser-trace/`).

## Errors

- Must not leave phase on `extension_connected` after a recording status.
- Must not keep `recording.active=false` when status said recording.

## Exit Code

- Not asserted.

```go
import (
	"net/http"
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
	assertPhase(t, resp, "recording")
	assertExtensionConnected(t, resp, true)
	assertSupportsBrowserTrace(t, resp, true)
	assertRecording(t, resp, true, req.EntryCount)

	if req.WindowID != 0 && resp.WindowID != req.WindowID {
		t.Fatalf("recording.window_id = %d, want %d", resp.WindowID, req.WindowID)
	}
	if resp.ExtensionVersion != req.HelloVersion {
		t.Fatalf("extension.version = %q, want %q", resp.ExtensionVersion, req.HelloVersion)
	}
}
```
