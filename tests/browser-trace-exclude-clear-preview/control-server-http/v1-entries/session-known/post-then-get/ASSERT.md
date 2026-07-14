## Expected

Requirement scenario **#3** — POST then GET:

- Final `GET /v1/entries` returns **HTTP 200**.
- JSON content-type (when present).
- `count` is **2** (or `entries` length 2).
- Both fixture URLs appear:
  - `https://api.example.com/v1/alpha`
  - `https://cdn.example.com/assets/app.js`
- Stage POST itself should be 2xx when implementer is complete (soft check).

## Side Effects

- Server stores last POST snapshot in per-session preview memory.
- No final HAR file required for this surface.

## Errors

- 404 on live session, empty body, or missing URLs are failures.
- Transport error from harness is a failure.

## Exit Code

- Not asserted (session cancelled after probe).

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
	assertJSONContentType(t, resp)
	assertEntryURLsMatch(t, resp, sampleEntryURLs())
	if resp.EntriesCount != 2 && len(resp.Entries) != 2 {
		t.Fatalf("count=%d len(entries)=%d, want 2; body=%s",
			resp.EntriesCount, len(resp.Entries), truncate(resp.BodyString, 400))
	}
	if resp.StagePostStatus != 0 && (resp.StagePostStatus < 200 || resp.StagePostStatus >= 300) {
		t.Fatalf("stage POST /v1/entries status=%d, want 2xx", resp.StagePostStatus)
	}
}
```
