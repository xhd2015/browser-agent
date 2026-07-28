## Expected

- `background.js` found.
- **register** path starts HTTP poll transport:
  - named starter (`startHttpPoll` / `startPollLoop` / …), **or**
  - register + `/v1/ext/hello` or `/v1/ext/poll` + transport/poll/http language.
- Must not be only `register` → `new WebSocket` without poll path.

## Side Effects

- None (read-only FS).

## Errors

- Register that only opens WebSocket fails this leaf.

## Exit Code

- Not asserted.

```go
import (
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	src := assertBackgroundPresent(t, req, resp)
	if !hasRegisterStartsPoll(src) {
		t.Fatalf("register must start HTTP poll transport (startHttpPoll/poll path), not WS-only connect; path=%v snippet=%s",
			resp.FoundPaths, truncate(src, 800))
	}
	// Stronger: at least hello or poll path so "register" + comment "poll" cannot pass alone.
	if !hasExtHelloPath(src) && !hasExtPollPath(src) {
		t.Fatalf("register→poll requires /v1/ext/hello or /v1/ext/poll token; path=%v",
			resp.FoundPaths)
	}
}
```
