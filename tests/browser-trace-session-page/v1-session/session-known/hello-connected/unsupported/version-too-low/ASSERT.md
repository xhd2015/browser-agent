# Expected

Requirement scenario **#3** (version too low variant):

- HTTP 200 JSON.
- `extension.connected` is **true**.
- `extension.supports_browser_trace` is **false**.
- `extension.version` is `"1.1.0"`.
- `extension.features` may include `browser-trace` (echoed), but support stays false.
- `phase` is `extension_connected`.
- `hint` non-empty; mentions update / version / support.

## Side Effects

- None.

## Errors

- Must not treat feature presence alone as sufficient when version &lt; 1.2.0.

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

	if resp.ExtensionVersion != "1.1.0" {
		t.Fatalf("extension.version = %q, want 1.1.0", resp.ExtensionVersion)
	}
	assertRecording(t, resp, false, 0)
	assertHintNonEmpty(t, resp)
	h := strings.ToLower(resp.Hint)
	ok := strings.Contains(h, "update") ||
		strings.Contains(h, "support") ||
		strings.Contains(h, "upgrade") ||
		strings.Contains(h, "version") ||
		strings.Contains(h, "1.2")
	if !ok {
		t.Fatalf("hint should mention update/version/support; got %q", resp.Hint)
	}
}
```
