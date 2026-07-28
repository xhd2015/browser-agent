## Expected

- HTTP status **200**.
- JSON `ok` is true.
- `phase` is **`extension_connected`**.
- `GET /v1/session` reports **extension.connected=true**.

## Side Effects

- Session is attach-ready for jobs (fastFailNoExtension passes).

## Errors

- 404 (route missing) or phase/connected false fails.

## Exit Code

- 0.

```go
import (
	"net/http"
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	if resp == nil {
		t.Fatal("resp is nil")
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status=%d want 200; body=%s (route POST /v1/ext/hello must exist)",
			resp.StatusCode, truncate(resp.BodyString, 400))
	}
	if !resp.HTTPOK {
		t.Fatalf("ok field false or missing; body=%s", truncate(resp.BodyString, 400))
	}
	if resp.Phase != "extension_connected" {
		t.Fatalf("phase=%q want extension_connected; body=%s",
			resp.Phase, truncate(resp.BodyString, 400))
	}
	if !resp.ExtensionConnected {
		t.Fatalf("extension.connected=false after hello; probe status=%d url=%s",
			resp.SessionProbeStatus, resp.SessionProbeURL)
	}
	assertJSONContentType(t, resp.ContentType)
	assertExitZero(t, resp)
}
```
