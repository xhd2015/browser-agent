## Expected

- `EmbedCompleteFS` on `testdata/session-page-html-only` with kind
  `session-page` returns **false** (HTML alone is not enough).

## Side Effects

- None (read-only).

## Errors

- true (incorrectly complete) fails this leaf.

## Exit Code

- 0.

```go
import (
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	assertCompleteFalse(t, resp, KindSessionPage)
	assertExitZero(t, resp)
}
```
