## Expected

Requirement scenario **#5** — GET `/preview` with entries:

- **HTTP 200**, HTML content type (or HTML-shaped body).
- Live preview markers: references `/v1/entries` poll and/or
  `data-browser-trace-preview` / preview root id / similar.
- At least one of the fixture entry URLs appears in the HTML **or**
  (poll-only) the live session id appears together with a poll marker
  (server may stream entries only via JS fetch).
- Soft: stage POST 2xx.

## Side Effects

- Preview page is served from control server memory (no disk snapshot).

## Errors

- 404 on live session, empty non-HTML body, or no live markers are failures.
- Missing both fixture URLs **and** (poll marker + session id) is a failure.

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
	body := resp.BodyString
	assertPreviewLiveMarkers(t, body)

	// Prefer seeing captured URL(s) in HTML; allow poll-only + session id.
	hasURL := false
	for _, u := range sampleEntryURLs() {
		if strings.Contains(body, u) {
			hasURL = true
			break
		}
	}
	sessionID := resp.RealSessionID
	if sessionID == "" {
		sessionID = req.SessionSuffix
	}
	hasSession := sessionID != "" && strings.Contains(body, sessionID)
	if !hasURL && !hasSession {
		t.Fatalf("preview HTML has neither fixture entry URL nor session id %q; body=%s",
			sessionID, truncate(body, 600))
	}
	if resp.StagePostStatus != 0 && (resp.StagePostStatus < 200 || resp.StagePostStatus >= 300) {
		t.Fatalf("stage POST /v1/entries status=%d, want 2xx", resp.StagePostStatus)
	}
}
```
