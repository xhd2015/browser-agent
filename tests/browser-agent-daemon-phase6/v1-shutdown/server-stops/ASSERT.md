## Expected

- `POST /v1/shutdown` returns **202**.
- `RunDaemon` goroutine exits after shutdown POST.
- `/v1/health` becomes unreachable (connection refused / down).
- `server.json` is **absent** after shutdown completes.

## Side Effects

- Removes `{BaseDir}/server.json` on clean shutdown.

## Errors

- Daemon still running or meta still present fails.

## Exit Code

- Not asserted.

```go
import (
	"net/http"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	if resp == nil {
		t.Fatal("resp is nil")
	}
	assertHTTPStatus(t, resp, http.StatusAccepted)
	if !resp.DaemonExited {
		t.Fatal("RunDaemon did not exit after POST /v1/shutdown")
	}
	if !resp.HealthDown {
		t.Fatal("health still reachable after shutdown")
	}
	if resp.DaemonMetaExists {
		t.Fatalf("server.json still exists at %s after shutdown", resp.DaemonMetaPath)
	}
}
```