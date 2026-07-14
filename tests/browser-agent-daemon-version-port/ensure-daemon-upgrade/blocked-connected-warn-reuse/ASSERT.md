## Expected

- client > daemon + connected → warn `cannot upgrade`; reuse; new session created

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
	assertContainsFold(t, resp.Stderr, "cannot upgrade", req.SessionIDA)
	if resp.KillFnCalled {
		t.Fatal("connected session must block kill")
	}
	if !resp.SessionCreated {
		t.Fatal("SessionNew should still create session B after blocked upgrade")
	}
	if resp.NewPID != 0 && resp.NewPID != resp.OldPID {
		t.Fatal("original daemon PID should remain")
	}
}
```
