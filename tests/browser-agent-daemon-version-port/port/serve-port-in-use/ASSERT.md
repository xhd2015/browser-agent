## Expected

- Second `RunDaemon` on same port fails non-zero
- Error mentions port in use / bind

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
	if resp.CLIErr == "" {
		t.Fatal("expected bind error when port in use")
	}
	if !resp.PortInUseSeen {
		t.Fatalf("error should mention port in use; err=%q", resp.CLIErr)
	}
}
```
