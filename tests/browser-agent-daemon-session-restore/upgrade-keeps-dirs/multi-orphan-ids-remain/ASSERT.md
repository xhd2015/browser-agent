## Expected

- Upgrade helper nil.
- Both `sess-o1` and `sess-o2` dirs exist with intact markers.

## Side Effects

- No orphan dir deleted.

## Errors

- Any missing dir/marker fails this leaf.

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
	for _, id := range []string{"sess-o1", "sess-o2"} {
		if !resp.OrphanDirExists[id] {
			t.Fatalf("orphan session dir removed for %s", id)
		}
		if !resp.MarkerIntact[id] {
			t.Fatalf("preserved.txt missing or rewritten for %s", id)
		}
	}
	assertExitZero(t, resp)
}
```
