## Expected

- `server.json` includes non-empty `daemon_version` matching daemon config

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
	if resp.MetaDaemonVer == "" {
		t.Fatal("server.json missing daemon_version")
	}
	if resp.MetaDaemonVer != req.DaemonVersion {
		t.Fatalf("daemon_version=%q want %q", resp.MetaDaemonVer, req.DaemonVersion)
	}
}
```
