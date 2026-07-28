## Expected

- PrepareUpgradeFn was called (and returned error internally).
- EnsureDaemon still succeeds.
- KillFn and SpawnFn still called.
- Stderr still has upgraded line.

## Side Effects

- None beyond upgrade path.

## Errors

- Aborting entire EnsureDaemon on prepare error fails this leaf.

## Exit Code

- 0.

```go
import (
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	if resp == nil {
		t.Fatal("resp is nil")
	}
	if resp.EnsureErr != nil {
		t.Fatalf("EnsureDaemon err=%v want nil (prepare is best-effort)", resp.EnsureErr)
	}
	if !resp.PrepareFnCalled {
		t.Fatal("PrepareUpgradeFn should still be invoked")
	}
	if !resp.KillFnCalled || !resp.SpawnFnCalled {
		t.Fatal("kill+spawn required after prepare failure")
	}
	assertContainsFold(t, resp.Stderr, "upgraded daemon", req.DaemonVersion, req.ClientVersion)
	assertExitZero(t, resp)
}
```
