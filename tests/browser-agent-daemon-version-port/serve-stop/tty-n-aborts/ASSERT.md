## Expected

- TTY + `n` → exit 0; daemon stays healthy; meta present

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
	if resp.DaemonStillUp != true {
		t.Fatal("daemon should remain running after n")
	}
	if !resp.MetaExists {
		t.Fatal("server.json should remain after abort")
	}
}
```
