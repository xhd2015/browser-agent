## Expected

- `EnsureDaemon` returns **nil** error.
- `SpawnFnCalled == true`.
- `DaemonMeta.PID` > 0; `Addr` non-empty; `BaseDir` equals request base dir.
- `GET /v1/health` OK at returned base URL.
- `server.json` exists under `{BaseDir}` after ensure.

## Side Effects

- New daemon process started via injected `SpawnFn`.

## Errors

- Spawn not called, health timeout, or missing meta fails.

## Exit Code

- Not asserted.

```go
import (
	"os"
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	if resp == nil {
		t.Fatal("resp is nil")
	}
	if resp.EnsureErr != "" {
		t.Fatalf("EnsureDaemon error: %s", resp.EnsureErr)
	}
	if !resp.SpawnFnCalled {
		t.Fatal("SpawnFn was not called; want spawn when daemon down")
	}
	meta := resp.DaemonMeta
	if meta.PID <= 0 {
		t.Fatalf("PID=%d want >0", meta.PID)
	}
	if strings.TrimSpace(meta.Addr) == "" {
		t.Fatal("Addr is empty")
	}
	if meta.BaseDir != req.BaseDir {
		t.Fatalf("BaseDir=%q want %q", meta.BaseDir, req.BaseDir)
	}
	baseURL := resp.BaseURL
	if baseURL == "" && meta.Addr != "" {
		baseURL = "http://" + meta.Addr
	}
	if !healthOK(baseURL) {
		t.Fatalf("daemon not healthy after spawn at %s", baseURL)
	}
	metaPath := daemonMetaPath(req.BaseDir)
	if _, err := os.Stat(metaPath); err != nil {
		t.Fatalf("server.json missing at %s: %v", metaPath, err)
	}
}```
