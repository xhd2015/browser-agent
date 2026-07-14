## Expected

Requirement **D2**:

- HTTP 200 HTML.
- Install guidance present when not connected — at least one of:
  - `chrome://extensions`
  - `load unpacked` / `Load unpacked`
  - stable marker `data-browser-agent-install` / `browser-agent-install` /
    `InstallGuideline`
- Prefer also mentioning Developer mode when full guideline is embedded.

## Side Effects

- None.

## Errors

- Missing all install cues fails product install UX.

## Exit Code

- Not asserted.

```go
import (
	"net/http"
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	assertHTTPStatus(t, resp, http.StatusOK)
	assertHTMLContentType(t, resp)
	body := resp.BodyString
	low := strings.ToLower(body)

	hasChromeExt := strings.Contains(low, "chrome://extensions")
	hasLoadUnpacked := strings.Contains(low, "load unpacked")
	hasMarker := strings.Contains(low, "data-browser-agent-install") ||
		strings.Contains(low, "browser-agent-install") ||
		strings.Contains(low, "installguideline") ||
		strings.Contains(low, "install-guideline")

	if !hasChromeExt && !hasLoadUnpacked && !hasMarker {
		t.Fatalf("HTML missing install guideline markers (chrome://extensions / load unpacked / install marker); body=%s",
			truncate(body, 700))
	}
}
```
