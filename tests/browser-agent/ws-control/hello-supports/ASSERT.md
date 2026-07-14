## Expected

Requirement **D1**:

- WS hello succeeded (`WSHelloOK`).
- GET /v1/session HTTP 200.
- `extension.connected` true.
- `supports_browser_agent` true.
- Version echoed when present.

## Side Effects

- Session phase may be `extension_connected` or similar (soft).

## Errors

- connected false after successful hello fails the leaf.

## Exit Code

- Not asserted.

```go
import (
	"net/http"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	if resp == nil {
		t.Fatal("resp is nil")
	}
	if !resp.WSHelloOK {
		t.Fatal("WSHelloOK=false")
	}
	assertHTTPStatus(t, resp, http.StatusOK)
	if !resp.ExtensionConnected {
		t.Fatalf("extension.connected=false after hello; body=%s", truncate(resp.BodyString, 400))
	}
	if !resp.SupportsBrowserAgent {
		t.Fatalf("supports_browser_agent=false; version=%q features=%v body=%s",
			resp.ExtensionVersion, resp.ExtensionFeatures, truncate(resp.BodyString, 400))
	}
}
```
