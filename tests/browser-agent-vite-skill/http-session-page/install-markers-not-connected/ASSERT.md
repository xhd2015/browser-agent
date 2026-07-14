## Expected

Requirement **A4**:

- HTTP 200 HTML.
- Install guidance present when not connected — at least one of:
  - `chrome://extensions`
  - `load unpacked` / `Load unpacked`
  - stable marker `data-browser-agent-install` / `browser-agent-install` /
    `InstallGuideline`

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
	if strings.TrimSpace(body) == "" {
		t.Fatal("HTML body empty")
	}
	if !hasInstallMarkers(body) {
		t.Fatalf("HTML missing install guideline markers (chrome://extensions / load unpacked / install marker); body=%s",
			truncate(body, 700))
	}
}
```
