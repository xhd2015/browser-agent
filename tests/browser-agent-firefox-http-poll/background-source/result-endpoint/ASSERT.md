## Expected

- `background.js` found.
- Path token **`/v1/ext/result`** present for HTTP job completion.

## Side Effects

- None (read-only FS).

## Errors

- Jobs handled only via WS `type: "result"` without HTTP result path fails this leaf.

## Exit Code

- Not asserted.

```go
import (
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	src := assertBackgroundPresent(t, req, resp)
	if !hasExtResultPath(src) {
		t.Fatalf("background missing /v1/ext/result path; path=%v snippet=%s",
			resp.FoundPaths, truncate(src, 700))
	}
}
```
