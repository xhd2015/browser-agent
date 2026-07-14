## Expected

- `0.2.0-beta` orders **below** `0.2.0` (compare = -1)

## Side Effects

- See leaf scenario (may mutate daemon meta, session dirs, or stderr).

## Errors

- Wrong version/port/upgrade/stop behavior fails the assertion.

## Exit Code

- Not asserted unless noted in Expected.

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	if resp.CompareResult != -1 {
		t.Fatalf("prerelease compare = %d, want -1 (beta < release)", resp.CompareResult)
	}
}
```
