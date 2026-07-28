## Expected

- `background.js` found.
- HTTP client surface present: **`fetch(`** (preferred) **or** explicit **`POST`** method.
- Coupled to poll transport language: poll/hello path tokens and/or poll-loop naming
  (`httpPoll`, `pollLoop`, `startHttpPoll`, etc.).

## Side Effects

- None (read-only FS).

## Errors

- `popup.js` health `fetch` alone is insufficient if background has no poll client
  (this leaf only reads background.js).
- WebSocket-only background without fetch/POST poll client fails.

## Exit Code

- Not asserted.

```go
import (
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	src := assertBackgroundPresent(t, req, resp)
	if !hasHttpPollFetch(src) {
		t.Fatalf("background missing fetch/HTTP POST client for poll transport; path=%v snippet=%s",
			resp.FoundPaths, truncate(src, 800))
	}
	// Prefer full endpoint set; soft nudge if only partial language.
	if !hasFetchClient(src) && !hasHttpPostMethod(src) {
		t.Fatalf("background must call fetch( or set method POST for HTTP poll; path=%v",
			resp.FoundPaths)
	}
}
```
