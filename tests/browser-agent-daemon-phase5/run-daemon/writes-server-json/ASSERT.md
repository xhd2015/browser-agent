## Expected

- `server.json` exists under `BaseDir`.
- `ReadDaemonMeta` fields match bound addr and `BaseDir`.
- `pid` equals `os.Getpid()`.
- `started_at` is non-zero.

## Side Effects

- Creates `{BaseDir}/server.json`.

## Errors

- Missing file or field mismatch fails.

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
	if !resp.DaemonMetaExists {
		t.Fatalf("server.json missing at %s; readErr=%q", resp.DaemonMetaPath, resp.ReadMetaErr)
	}
	if resp.ReadMetaErr != "" {
		t.Fatalf("ReadDaemonMeta err=%s", resp.ReadMetaErr)
	}
	daemonMetaFieldsMatch(t, resp.DaemonMeta, resp.Addr, req.BaseDir)
	if len(resp.DaemonMetaRaw) > 0 && !strings.HasSuffix(string(resp.DaemonMetaRaw), "\n") {
		t.Fatalf("server.json must end with trailing newline")
	}
}
```