## Expected

- Phase 3 policy: client > daemon + connected → **upgrade proceeds** (no Q1 block).
- SessionNew still creates session B after upgrade.
- Stderr must **not** say `cannot upgrade` (connected no longer blocks).
- Stderr should mention upgraded daemon (and may warn still-waiting reattach).

## Side Effects

- Old daemon may be killed and respawned; session dirs kept (Phase 1).
- Fake extension on session A may not reattach within wait (soft warn).

## Errors

- Reuse-without-upgrade (old Q1 `cannot upgrade`) fails this leaf.

## Exit Code

- Not asserted unless noted in Expected.

```go
import (
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	// Phase 3: connected sessions no longer block EnsureDaemon upgrade.
	assertNotContainsFold(t, resp.Stderr, "cannot upgrade")
	assertContainsFold(t, resp.Stderr, "upgraded daemon")
	if !resp.SessionCreated {
		t.Fatal("SessionNew should still create session B after upgrade with connected sessions")
	}
}
```
