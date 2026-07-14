## Expected

Requirement scenarios **#1** and **#5** — `/go` HTML while not connected:

- HTTP 200, HTML content type.
- Body non-empty; includes session id.
- Install panel marker present (always-visible contract).
- Panel **expanded**: `open` on panel details and/or `data-default-open="true"`.
- Panel is **not** `display:none` in server HTML.
- Body contains `chrome://extensions` as text.
- Body contains path guidance (`/extension/…`, BaseDir, or data-*-path attrs).

## Side Effects

- Extract under BaseDir from session start.

## Errors

- Missing panel when not connected is a failure.
- Collapsed panel when not connected is a failure.
- Relying only on `<a href="chrome://…">` without text is insufficient.

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
	assertHTMLContentType(t, resp)
	assertSessionIDInBody(t, req, resp)

	body := resp.BodyString
	assertInstallPanelPresent(t, body)
	assertPanelNotDisplayNone(t, body)
	assertPanelExpanded(t, body)
	assertInstallGuidance(t, req, body)
}
```
