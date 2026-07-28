## Expected

- `background.js` found.
- Path token **`/v1/ext/poll`** present.
- Long-poll wait parameter **`wait_ms`** or **`waitMs`** present.

## Side Effects

- None (read-only FS).

## Errors

- Missing poll path or wait_ms fails (server defaults exist, but client must send wait_ms).

## Exit Code

- Not asserted.

```go
import (
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	src := assertBackgroundPresent(t, req, resp)
	if !hasExtPollPath(src) {
		t.Fatalf("background missing /v1/ext/poll path; path=%v snippet=%s",
			resp.FoundPaths, truncate(src, 700))
	}
	if !hasWaitMs(src) {
		t.Fatalf("background poll must send wait_ms (or waitMs); path=%v snippet=%s",
			resp.FoundPaths, truncate(src, 700))
	}
}
```
