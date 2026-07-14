## Expected

- `QueryDaemonStatus` returns **nil** error.
- `Running=false` for dead pid in `server.json`.
- `server.json` bytes **unchanged** after query.

## Side Effects

- Stale meta file remains (not removed by status query).

## Errors

- Query error, `Running=true`, or meta mutation fails.

## Exit Code

- Not asserted.

```go
import (
	"testing"

	"github.com/xhd2015/browser-agent/browseragent"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	if resp == nil {
		t.Fatal("resp is nil")
	}
	if resp.QueryErr != "" {
		t.Fatalf("QueryDaemonStatus error: %s", resp.QueryErr)
	}
	if resp.Status.Running {
		t.Fatal("Running=true, want false for stale pid")
	}
	if !resp.DaemonMetaBeforeHit || !resp.DaemonMetaAfterHit {
		t.Fatal("server.json missing before/after stale status query")
	}
	assertMetaUnchanged(t, resp)
	stalePID := req.StalePID
	if stalePID == 0 {
		stalePID = 999999999
	}
	if !browseragent.IsProcessAlive(stalePID) {
		return
	}
	t.Fatalf("StalePID=%d is unexpectedly alive; pick a dead pid for this leaf", stalePID)
}
```