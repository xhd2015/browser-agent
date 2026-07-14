## Expected

After implementer lands session list (**RED** on current code):

- Empty `BaseDir` with no `server.json`.
- `HandleCLI` returns nil; `ExitCode` 0.
- Stderr contains **`warning:`** and **`daemon not running`** (case-insensitive ok).
- Stderr mentions `BaseDir` path (or basename).
- Stdout is empty table / `0 sessions` (no session rows).

## Side Effects

- Read-only; no daemon spawn.

## Errors

- Non-zero exit, missing warning, or error instead of graceful empty output fails.

## Exit Code

- 0.

```go
import (
	"path/filepath"
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	if resp == nil {
		t.Fatal("resp is nil")
	}
	if resp.DispatchTimedOut {
		t.Fatal("session list timed out")
	}
	if resp.DaemonMetaExists {
		t.Fatal("server.json should be absent for daemon-down leaf")
	}

	assertExitZero(t, resp)
	assertContainsFold(t, resp.Stderr, "warning:", "daemon not running")

	lowErr := strings.ToLower(resp.Stderr)
	if !strings.Contains(lowErr, strings.ToLower(req.BaseDir)) &&
		!strings.Contains(lowErr, strings.ToLower(filepath.Base(req.BaseDir))) {
		t.Fatalf("stderr should mention base-dir %q; stderr=%q", req.BaseDir, truncate(resp.Stderr, 600))
	}

	lowOut := strings.ToLower(resp.Stdout)
	if strings.Contains(lowOut, "sess-") {
		t.Fatalf("stdout must not list sessions when daemon down; stdout=%q", truncate(resp.Stdout, 400))
	}
	if resp.Stdout != "" {
		if !strings.Contains(lowOut, "0 sessions") && !strings.Contains(lowOut, "session id") {
			t.Fatalf("stdout should be empty or zero-session summary; stdout=%q", truncate(resp.Stdout, 400))
		}
	}
}
```
