# Expected

Requirement scenario **#2** — hello with version + features including
`browser-trace`, version ≥ 1.2.0:

- HTTP 200 JSON.
- `session_id` matches live session.
- `phase` is `extension_connected`.
- `extension.connected` is **true**.
- `extension.supports_browser_trace` is **true**.
- `extension.version` equals the hello version (`1.2.0`).
- `extension.features` includes `browser-trace`.
- `recording.active` is false; entry_count 0.
- Hint may be empty or guide user to start browsing; if non-empty, must not claim
  “extension not installed” as the primary story (connection succeeded).

## Side Effects

- None asserted beyond probe correctness.

## Errors

- Must not leave `supports_browser_trace` false when both gates pass.
- Must not stay on `waiting_extension` after a successful hello.

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
	assertPhase(t, resp, "extension_connected")
	assertExtensionConnected(t, resp, true)
	assertSupportsBrowserTrace(t, resp, true)

	if resp.ExtensionVersion != req.HelloVersion {
		t.Fatalf("extension.version = %q, want %q", resp.ExtensionVersion, req.HelloVersion)
	}
	if !featuresContain(resp.ExtensionFeatures, "browser-trace") {
		t.Fatalf("extension.features %v missing %q", resp.ExtensionFeatures, "browser-trace")
	}
	assertRecording(t, resp, false, 0)
	assertReadyCountdownPresent(t, resp)
}
```
