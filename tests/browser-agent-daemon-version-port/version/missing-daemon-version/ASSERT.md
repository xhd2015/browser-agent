## Expected

- Empty daemon version normalizes to `0.0.0`
- Compare vs client `0.2.0` is **+1** (client newer)

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
	if resp.ClientVersion != "0.0.0" {
		t.Fatalf("EffectiveDaemonVersion empty = %q, want 0.0.0", resp.ClientVersion)
	}
	if resp.CompareResult != 1 {
		t.Fatalf("client newer than 0.0.0: compare=%d want 1", resp.CompareResult)
	}
}
```
