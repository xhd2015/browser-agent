## Expected

- Disconnected orphan session → stderr lists id; session dir **kept** (Phase 1 restore policy)
- Warning still emitted; dirs preserved so respawned daemon can RestoreSessionsFromDisk

## Side Effects

- See leaf scenario (may mutate daemon meta, session dirs, or stderr).
- Orphan session directories are **not** wiped on normal upgrade (policy change from prior wipe-on-upgrade).

## Errors

- Wrong version/port/upgrade/stop behavior fails the assertion.

## Exit Code

- Not asserted unless noted in Expected.

```go
import (
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	assertContainsFold(t, resp.Stderr, "orphan", req.OrphanID)
	// Phase 1 product policy: keep orphan session dirs across upgrade so the
	// respawned daemon can restore them from disk.
	if !resp.OrphanDirExists {
		t.Fatalf("orphan dir removed for %s; upgrade must keep dirs for restore", req.OrphanID)
	}
}
```
