## Expected

Requirement **E2**:

- HTTP 200.
- `extension.connected` true.
- `supports_browser_agent` true.

## Side Effects

- None.

## Errors

- supports false with valid hello fails capability gate.

## Exit Code

- Not asserted.

```go
import (
	"net/http"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	assertHTTPStatus(t, resp, http.StatusOK)
	if !resp.WSHelloOK {
		t.Fatal("WSHelloOK=false")
	}
	if !resp.ExtensionConnected {
		t.Fatalf("connected=false after hello; body=%s", truncate(resp.BodyString, 400))
	}
	if !resp.SupportsBrowserAgent {
		t.Fatalf("supports_browser_agent=false; body=%s", truncate(resp.BodyString, 400))
	}
}
```
