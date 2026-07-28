## Expected

- EnsureDaemon succeeds (err nil).
- KillFn and SpawnFn were called.
- Stderr does **not** contain the old block phrase `cannot upgrade`.
- Stderr contains upgraded line with old and new versions.

## Side Effects

- Fake health version advances to client version after spawn.

## Errors

- Reuse-without-kill (old Q1) fails this leaf.

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
		t.Fatalf("EnsureDaemon err=%v want nil (connected must not block upgrade)", resp.EnsureErr)
	}
	if !resp.KillFnCalled {
		t.Fatal("KillFn not called; connected sessions must no longer block upgrade")
	}
	if !resp.SpawnFnCalled {
		t.Fatal("SpawnFn not called after kill")
	}
	assertNotContainsFold(t, resp.Stderr, "cannot upgrade")
	assertContainsFold(t, resp.Stderr, "upgraded daemon", req.DaemonVersion, req.ClientVersion)
	assertExitZero(t, resp)
}
```
