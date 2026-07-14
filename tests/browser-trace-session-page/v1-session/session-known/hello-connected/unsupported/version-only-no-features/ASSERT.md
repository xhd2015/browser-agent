# Expected

Documents product rule: **version alone is not enough**.

- HTTP 200 JSON.
- `extension.connected` is **true** (hello accepted).
- `extension.supports_browser_trace` is **false**.
- `extension.version` is `"1.2.0"`.
- `extension.features` is empty or absent.
- `phase` is `extension_connected`.
- `hint` non-empty; mentions update / support / feature.

## Side Effects

- None.

## Errors

- Must not grant supports_browser_trace=true when features are omitted,
  even if version ≥ 1.2.0.

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
		t.Fatalf("features should not invent browser-trace when omitted on hello; got %v",
			resp.ExtensionFeatures)
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
		t.Fatalf("hint should mention update/support/feature; got %q", resp.Hint)
	}
}
```
