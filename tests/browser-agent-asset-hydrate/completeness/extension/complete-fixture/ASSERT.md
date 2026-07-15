## Expected

- `EmbedCompleteFS` on `testdata/extension-complete` with kind `extension`
  returns **true**.

## Side Effects

- None (read-only).

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
	assertCompleteTrue(t, resp, KindExtension)
	assertExitZero(t, resp)
}
```
