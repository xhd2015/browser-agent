## Expected

- `HandleCLI serve --status` returns **nil** (exit **0**).
- Stdout contains formatted status markers (`status`, `sessions`, table headers).
- Stdout mentions created session id.
- Daemon remains healthy; `server.json` unchanged.

## Side Effects

- No new daemon spawn from `--status`.

## Errors

- Non-nil CLI error, missing stdout markers, or meta mutation fails.

## Exit Code

- **0** (`HandleCLI` returns nil).

```go
import (
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
	assertContainsFold(t, resp.Stdout, "status", "sessions")
	assertContainsFold(t, resp.Stdout, "pid", "addr", "session")
	assertContainsFold(t, resp.Stdout, req.SessionID)
	assertMetaUnchanged(t, resp)
	if !resp.DaemonHealthyAfter {
		t.Fatal("daemon not healthy after serve --status")
	}
}
```