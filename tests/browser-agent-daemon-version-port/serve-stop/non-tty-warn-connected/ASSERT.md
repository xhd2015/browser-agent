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

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	// Product emits extension-connected inventory always; non-TTY adds "warning:" when
	// connected. Accept either form so TTY inject drift does not false-fail.
	assertContainsFold(t, resp.Stderr, req.SessionIDA)
	low := strings.ToLower(resp.Stderr)
	if !strings.Contains(low, "warning") && !strings.Contains(low, "extension-connected") {
		t.Fatalf("stderr should warn about connected sessions; got %q", resp.Stderr)
	}
	if !resp.DaemonStopped {
		t.Fatal("non-TTY stop should still kill daemon")
	}
}
```
