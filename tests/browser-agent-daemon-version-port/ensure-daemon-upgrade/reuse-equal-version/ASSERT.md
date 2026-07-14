## Expected

- Equal versions → reuse daemon; `KillFn`/`SpawnFn` not called; no upgrade warn

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
	if resp.KillFnCalled || resp.SpawnFnCalled {
		t.Fatal("equal version must reuse without kill/spawn")
	}
	if resp.NewPID != 0 && resp.NewPID != resp.OldPID {
		t.Fatalf("unexpected PID change old=%d new=%d", resp.OldPID, resp.NewPID)
	}
	assertNotContainsFold(t, resp.Stderr, "cannot upgrade", "upgrading daemon")
}
```
