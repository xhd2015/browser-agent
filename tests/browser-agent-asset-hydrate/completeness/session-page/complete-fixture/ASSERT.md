## Expected

- `EmbedCompleteFS` on `testdata/session-page-complete` with kind
  `session-page` returns **true**.
- Response `Complete` is true; FSRoot points at the fixture directory.

## Side Effects

- None (read-only FS inspection).

## Errors

- false / missing API fails this leaf.

## Exit Code

- 0.

```go
import (
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	assertCompleteTrue(t, resp, KindSessionPage)
	assertExitZero(t, resp)
}
```
