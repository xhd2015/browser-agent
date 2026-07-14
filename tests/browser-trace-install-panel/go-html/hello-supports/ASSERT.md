## Expected

Requirement scenario **#2** — `/go` after hello **with** supports:

- HTTP 200, HTML content type; session id present.
- Install panel **still present** (this is the always-visible regression vs old
  `showInstall := !helloOK || !supportsBT` omit).
- Panel **collapsed** by default: no `open` on panel details and/or
  `data-default-open="false"` (must not claim expanded via default-open true).
- Panel is **not** `display:none` in server HTML (collapse ≠ remove/hide whole panel).
- Summary / install markers remain so the user can expand to update/reload.

## Side Effects

- Hello stages connected + supports_browser_trace=true before HTML serve.

## Errors

- Missing panel when connected+supports is the primary failure mode of the old UX.
- Expanded-by-default when fully working is a failure (unless product only uses
  client-side collapse — server should still set collapsed default at serve time).

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
	assertPanelCollapsed(t, body)

	// Markers / summary affordance should remain for manual expand.
	low := strings.ToLower(body)
	hasSummaryOrHeading := strings.Contains(low, "<summary") ||
		strings.Contains(low, "install") ||
		strings.Contains(low, "extension")
	if !hasSummaryOrHeading {
		t.Fatalf("collapsed panel should still expose summary/install wording; body=%s",
			truncate(body, 600))
	}
}
```
