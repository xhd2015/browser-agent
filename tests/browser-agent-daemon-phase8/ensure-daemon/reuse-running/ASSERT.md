## Expected

- `EnsureDaemon` returns **nil** error.
- `SpawnFnCalled == false`.
- `DaemonMeta.Addr` equals live daemon addr; `PID` > 0; `BaseDir` equals request base dir.
- `server.json` readable; health still OK after ensure.

## Side Effects

- No new daemon spawn; existing daemon remains healthy.

## Errors

- Spawn called, wrong meta, or ensure error fails.

## Exit Code

- Not asserted.

```go
import (
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
	if resp.SpawnFnCalled {
		t.Fatal("SpawnFn was called; want reuse without spawn")
	}
	meta := resp.DaemonMeta
	if meta.PID <= 0 {
		t.Fatalf("PID=%d want >0", meta.PID)
	}
	if meta.Addr != resp.Addr {
		t.Fatalf("Addr=%q want %q", meta.Addr, resp.Addr)
	}
	if meta.BaseDir != req.BaseDir {
		t.Fatalf("BaseDir=%q want %q", meta.BaseDir, req.BaseDir)
	}
	wantBaseURL := strings.TrimRight("http://"+resp.Addr, "/")
	if strings.TrimRight(meta.BaseURL, "/") != wantBaseURL {
		t.Fatalf("BaseURL=%q want %q", meta.BaseURL, wantBaseURL)
	}
	if !healthOK(resp.BaseURL) {
		t.Fatal("daemon not healthy after EnsureDaemon reuse")
	}
}```
