## Expected

Requirement scenarios **#4** + **#5** — preview after empty push:

- **HTTP 200**, HTML content type (or HTML-shaped body).
- Live preview markers still present (page shell remains).
- Pre-clear fixture URLs should **not** appear as current embedded rows.
  (Poll-only pages that only contain JS may omit URLs entirely — that is OK
  if no fixture URL is embedded.)
- Empty-state signal recommended: wording like empty / no requests / 0 entries,
  or an empty table body — at least one soft signal when URLs are absent.

## Side Effects

- Clear POST replaces server snapshot; preview reads the empty snapshot.

## Errors

- 404 on live session is a failure.
- HTML that still hard-embeds both pre-clear fixture URLs as current data is a failure.

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

	// Must not hard-embed pre-clear fixture URLs after clear.
	// (If implementer SSR embeds entries, they must reflect empty snapshot.)
	embedded := 0
	for _, u := range sampleEntryURLs() {
		if strings.Contains(body, u) {
			embedded++
		}
	}
	if embedded == len(sampleEntryURLs()) {
		t.Fatalf("preview HTML still embeds all pre-clear fixture URLs after clear; body=%s",
			truncate(body, 600))
	}

	// Soft empty-state signal when no URLs embedded.
	if embedded == 0 {
		low := strings.ToLower(body)
		hasEmptyHint := strings.Contains(low, "empty") ||
			strings.Contains(low, "no request") ||
			strings.Contains(low, "no entries") ||
			strings.Contains(low, "0 entries") ||
			strings.Contains(low, "count\":0") ||
			strings.Contains(low, "count: 0") ||
			strings.Contains(low, "data-count=\"0\"") ||
			strings.Contains(low, "data-entry-count=\"0\"") ||
			strings.Contains(low, "<tbody></tbody>") ||
			strings.Contains(low, "nothing captured") ||
			strings.Contains(low, "cleared")
		// Poll-only shell without empty copy is acceptable if live markers exist
		// (already checked) — do not hard-fail missing empty wording.
		_ = hasEmptyHint
	}

	if resp.ClearPostStatus != 0 && (resp.ClearPostStatus < 200 || resp.ClearPostStatus >= 300) {
		t.Fatalf("clear POST /v1/entries status=%d, want 2xx", resp.ClearPostStatus)
	}
}
```
