## Expected

- `background.js` found under `Firefox-Ext-Browser-Agent`.
- Source contains path token **`/v1/ext/hello`** (HTTP poll hello attach).

## Side Effects

- None (read-only FS).

## Errors

- WS-only background without `/v1/ext/hello` fails.

## Exit Code

- Not asserted.

```go
import (
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	src := assertBackgroundPresent(t, req, resp)
	if !hasExtHelloPath(src) {
		t.Fatalf("background missing /v1/ext/hello HTTP attach path; path=%v snippet=%s",
			resp.FoundPaths, truncate(src, 700))
	}
}
```
