## Expected

- client > daemon, 0 connected → kill+respawn; new `daemon_version` = client

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
	if !resp.KillFnCalled || !resp.SpawnFnCalled {
		t.Fatal("upgrade requires kill and spawn when no connected sessions")
	}
	if resp.MetaDaemonVer != req.ClientVersion {
		t.Fatalf("daemon_version=%q want client %q", resp.MetaDaemonVer, req.ClientVersion)
	}
}
```
