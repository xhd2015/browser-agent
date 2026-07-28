## Expected

- Upgrade helper returns nil (kill/spawn injected success).
- `SessionDirExists(baseDir, sess-orph01)` true.
- `preserved.txt` still contains `keep-me` (not recreated empty).

## Side Effects

- Dir must not be wiped by `removeSessionDirs`.

## Errors

- Missing dir or missing marker fails this leaf (current wipe behavior → RED).

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
	if resp.UpgradeErr != nil {
		t.Fatalf("ensureDaemonKillAndRespawn err=%v want nil", resp.UpgradeErr)
	}
	id := "sess-orph01"
	if !resp.OrphanDirExists[id] {
		t.Fatalf("orphan session dir removed for %s (upgrade must keep dirs)", id)
	}
	if !resp.MarkerIntact[id] {
		t.Fatalf("preserved.txt missing or rewritten for %s", id)
	}
	assertExitZero(t, resp)
}
```
