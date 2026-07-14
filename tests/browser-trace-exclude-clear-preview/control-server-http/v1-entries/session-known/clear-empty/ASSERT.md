## Expected

Requirement scenario **#4** — POST empty after clear:

- Final `GET /v1/entries` returns **HTTP 200**.
- `count` is **0** and `entries` is empty / absent-as-empty.
- Fixture URLs from the pre-clear POST must **not** remain in the GET body.
- Clear POST status should be 2xx when implementer is complete (soft check).

## Side Effects

- Server preview memory is replaced by the empty snapshot (not append).

## Errors

- Non-zero count or leftover URLs after clear is a failure.
- 404 on live session is a failure.

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
	assertJSONContentType(t, resp)

	if resp.EntriesCount != 0 {
		t.Fatalf("after clear, count=%d, want 0; body=%s",
			resp.EntriesCount, truncate(resp.BodyString, 400))
	}
	if len(resp.Entries) != 0 {
		t.Fatalf("after clear, len(entries)=%d, want 0; urls=%v",
			len(resp.Entries), resp.EntryURLs)
	}
	// Ensure previous fixture URLs are gone.
	for _, u := range sampleEntryURLs() {
		if strings.Contains(resp.BodyString, u) {
			t.Fatalf("after clear, GET body still contains %q; body=%s",
				u, truncate(resp.BodyString, 400))
		}
	}
	if resp.ClearPostStatus != 0 && (resp.ClearPostStatus < 200 || resp.ClearPostStatus >= 300) {
		t.Fatalf("clear POST /v1/entries status=%d, want 2xx", resp.ClearPostStatus)
	}
}
```
