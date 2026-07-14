## Expected

- `EnsureDaemon` spawn-when-down on explicit `--port` still works (phase8 parity)

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
	if !resp.SpawnFnCalled {
		t.Fatal("SpawnFn not called when daemon down")
	}
	if resp.Meta.PID <= 0 {
		t.Fatalf("PID=%d want >0", resp.Meta.PID)
	}
	if !healthOK(resp.BaseURL) {
		t.Fatalf("daemon not healthy at %s", resp.BaseURL)
	}
}
```
