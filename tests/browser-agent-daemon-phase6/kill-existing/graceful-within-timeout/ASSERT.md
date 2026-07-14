## Expected

- `KillExistingDaemon` returns **nil** (no error).
- `/v1/health` becomes unreachable.
- `RunDaemon` goroutine exits.
- `server.json` is **absent**.

## Side Effects

- Removes stale `{BaseDir}/server.json`.

## Errors

- Non-nil kill error or lingering meta fails.

## Exit Code

- Not asserted.

```go
import (
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	if resp == nil {
		t.Fatal("resp is nil")
	}
	if resp.KillErr != "" {
		t.Fatalf("KillExistingDaemon error: %s", resp.KillErr)
	}
	if !resp.HealthDown {
		t.Fatal("health still reachable after KillExistingDaemon")
	}
	if !resp.DaemonExited {
		t.Fatal("RunDaemon did not exit after KillExistingDaemon")
	}
	if resp.DaemonMetaExists {
		t.Fatalf("server.json still exists at %s after kill", resp.DaemonMetaPath)
	}
}
```