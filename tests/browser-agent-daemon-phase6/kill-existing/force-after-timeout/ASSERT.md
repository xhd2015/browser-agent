## Expected

- `KillExistingDaemon` returns **nil** despite short timeout (force path).
- `/v1/health` becomes unreachable within `ShutdownWait`.
- `server.json` is **absent** after kill completes.

## Side Effects

- Force-kills daemon pid when graceful drain exceeds timeout.
- Removes stale `{BaseDir}/server.json`.

## Errors

- Kill helper error or lingering meta fails.

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
		t.Fatal("health still reachable after force kill path")
	}
	if resp.DaemonMetaExists {
		t.Fatalf("server.json still exists at %s after force kill", resp.DaemonMetaPath)
	}
}
```