## Expected

- `QueryDaemonStatus` returns **nil** error.
- `Running=true`.
- `DaemonVersion` is non-empty and equals configured daemon version
  (`browseragent.ClientVersion()` default).
- `server.json` bytes **unchanged** after query.

## Side Effects

- Read-only probe; no meta mutation.

## Errors

- Query error, empty `DaemonVersion`, or meta mutation fails.

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
	st := resp.Status
	if !st.Running {
		t.Fatal("Running=false, want true")
	}
	gotVer := richStatusString(st, "DaemonVersion")
	if gotVer == "" {
		t.Fatal("DaemonVersion empty, want non-empty when daemon running")
	}
	wantVer := browseragent.EffectiveDaemonVersion(req.DaemonVersion)
	if gotVer != wantVer {
		t.Fatalf("DaemonVersion=%q want %q", gotVer, wantVer)
	}
	assertMetaUnchanged(t, resp)
	if !resp.DaemonHealthyAfter {
		t.Fatal("daemon not healthy after read-only status query")
	}
}
```