## Expected

- `CompareVersion("0.2.0", "0.1.0")` returns **+1**

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
	if resp.CompareResult != 1 {
		t.Fatalf("CompareVersion(0.2.0, 0.1.0) = %d, want 1", resp.CompareResult)
	}
}
```
