## Expected

- `background.js` found.
- **`prepare_reconnect`** (or `prepareReconnect`) handling present.
- Reachable from HTTP poll: **`/v1/ext/poll`** and/or poll **events** drain language
  and/or full hello|poll|result endpoint set.

## Side Effects

- None (read-only FS).

## Errors

- prepare_reconnect only on WS path with no poll transport fails when poll
  endpoints are absent (`hasPrepareReconnectOnPoll` false).

## Exit Code

- Not asserted.

```go
import (
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	src := assertBackgroundPresent(t, req, resp)
	if !hasPrepareReconnectOnPoll(src) {
		t.Fatalf("background must handle prepare_reconnect from poll events (need marker + poll/events path); path=%v snippet=%s",
			resp.FoundPaths, truncate(src, 800))
	}
	if !hasPrepareReconnectMarker(src) {
		t.Fatalf("missing prepare_reconnect marker; path=%v", resp.FoundPaths)
	}
}
```
