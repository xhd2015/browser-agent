## Expected

- non-TTY + connected → stderr warning with session id; daemon stops

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
	assertContainsFold(t, resp.Stderr, "warning", req.SessionIDA)
	if !resp.DaemonStopped {
		t.Fatal("non-TTY stop should still kill daemon")
	}
}
```
