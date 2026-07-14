## Expected

Requirement scenario **#3** — `/go` after hello **without** supports:

- HTTP 200, HTML content type; session id present.
- Install panel **still present** (always-visible).
- Panel **expanded** (`open` and/or `data-default-open="true"`).
- Not `display:none` on the panel root.
- Install markers remain; `chrome://extensions` and path guidance preferred
  (panel body still useful for update/reload).

## Side Effects

- Hello stages session connected + supports=false before HTML serve.

## Errors

- Omitting the panel because “something connected” is a failure.
- Collapsed panel when supports is false is a failure.

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
	assertHTMLContentType(t, resp)
	assertSessionIDInBody(t, req, resp)

	body := resp.BodyString
	assertInstallPanelPresent(t, body)
	assertPanelNotDisplayNone(t, body)
	assertPanelExpanded(t, body)

	// Soft-strong: keep chrome://extensions when panel body is rendered.
	if !strings.Contains(body, "chrome://extensions") {
		t.Fatalf("HTML should still include chrome://extensions in install panel; body=%s",
			truncate(body, 600))
	}
}
```
