## Expected

- `HandleCLI serve --status` returns **nil** (exit **0**).
- Stdout contains **not running** indication.
- Stdout contains **`Extension (embedded)`** block with **`version`**, **`md5`**, **`path`**.
- `server.json` still **absent** (no writes).

## Side Effects

- No daemon spawn; no meta file created.

## Errors

- Non-nil CLI error, missing extension block, or meta creation fails.

## Exit Code

- **0** (`HandleCLI` returns nil).

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
	if resp.CLIErr != "" {
		t.Fatalf("HandleCLI error: %s", resp.CLIErr)
	}
	if resp.Stdout == "" {
		t.Fatal("stdout is empty")
	}
	low := strings.ToLower(resp.Stdout)
	if !strings.Contains(low, "status") {
		t.Fatalf("stdout missing status marker; stdout=%q", truncate(resp.Stdout, 600))
	}
	notRunningMarkers := []string{"not running", "stopped", "running: false", "running:false", "running=false"}
	seen := false
	for _, m := range notRunningMarkers {
		if strings.Contains(low, strings.ToLower(m)) {
			seen = true
			break
		}
	}
	if !seen {
		t.Fatalf("stdout missing not-running marker; stdout=%q", truncate(resp.Stdout, 600))
	}
	assertContainsFold(t, resp.Stdout,
		"extension (embedded)", "version", "md5", "path",
	)
	if resp.DaemonMetaBeforeHit || resp.DaemonMetaAfterHit {
		t.Fatal("server.json should remain absent for not-running CLI status")
	}
	assertMetaUnchanged(t, resp)
}
```