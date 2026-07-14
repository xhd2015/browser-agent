## Expected

- `CompareVersion("0.2.0", "0.2.0")` returns **0**

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
	if resp.CompareResult != 0 {
		t.Fatalf("CompareVersion equal = %d, want 0", resp.CompareResult)
	}
}
```
