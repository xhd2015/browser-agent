## Expected

Explicit anti WS-only contract:

1. `background.js` found.
2. **`isWsOnlyTransport(src)` must be false** — all three HTTP poll routes present:
   - `/v1/ext/hello`
   - `/v1/ext/poll`
   - `/v1/ext/result`
3. WebSocket / `/v1/ws` may still appear (optional try-first / dual-mode).

## Side Effects

- None (read-only FS).

## Errors

- Phase-1 WS-only connect (hello over WebSocket only, no `/v1/ext/*`) **fails**.

## Exit Code

- Not asserted.

```go
import (
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	src := assertBackgroundPresent(t, req, resp)

	if isWsOnlyTransport(src) {
		t.Fatalf("background still WS-only (need /v1/ext/hello|poll|result HTTP poll transport; WS may remain); path=%v snippet=%s",
			resp.FoundPaths, truncate(src, 800))
	}
	if !hasHttpPollEndpoints(src) {
		t.Fatalf("HTTP poll endpoints incomplete: hello=%v poll=%v result=%v; path=%v",
			hasExtHelloPath(src), hasExtPollPath(src), hasExtResultPath(src), resp.FoundPaths)
	}
}
```
