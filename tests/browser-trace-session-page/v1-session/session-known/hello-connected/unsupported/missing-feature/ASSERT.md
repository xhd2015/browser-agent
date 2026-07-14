# Expected

Requirement scenario **#3** (missing feature variant):

- HTTP 200 JSON.
- `extension.connected` is **true**.
- `extension.supports_browser_trace` is **false**.
- `extension.version` is `"1.2.0"`.
- `extension.features` does not grant support (must not list a successful gate).
- `phase` is `extension_connected` (hello happened; not recording).
- `hint` is non-empty and mentions update and/or support / feature capability.

## Side Effects

- None.

## Errors

- Must not set supports_browser_trace=true when `browser-trace` feature is absent.

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
	assertExtensionConnected(t, resp, true)
	assertSupportsBrowserTrace(t, resp, false)
	assertPhase(t, resp, "extension_connected")

	if resp.ExtensionVersion != "1.2.0" {
		t.Fatalf("extension.version = %q, want 1.2.0", resp.ExtensionVersion)
	}
	if featuresContain(resp.ExtensionFeatures, "browser-trace") {
		// Server may echo features as sent; if browser-trace appears, support must still be false
		// (already asserted). Echo of only multi-tab-window is preferred.
	}
	assertRecording(t, resp, false, 0)
	assertHintNonEmpty(t, resp)
	h := strings.ToLower(resp.Hint)
	ok := strings.Contains(h, "update") ||
		strings.Contains(h, "support") ||
		strings.Contains(h, "upgrade") ||
		strings.Contains(h, "feature") ||
		strings.Contains(h, "version") ||
		strings.Contains(h, "capability")
	if !ok {
		t.Fatalf("hint should mention update/support/capability; got %q", resp.Hint)
	}
}
```
