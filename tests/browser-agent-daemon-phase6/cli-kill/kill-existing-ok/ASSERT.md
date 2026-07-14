## Expected

- First daemon health goes down after `serve --kill-existing`.
- Stderr contains operator-facing kill/shutdown status (case-insensitive).
- Second serve becomes healthy on same addr.
- `HandleCLI` returns **nil** (exit **0**) after harness shuts down second serve.

## Side Effects

- Removes first daemon `server.json`; second serve writes fresh meta until shutdown.

## Errors

- Missing stderr status, first daemon still healthy, or non-nil CLI error fails.

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
	if !resp.KillMessageSeen {
		t.Fatalf("stderr missing kill-existing status; stderr=%q stdout=%q",
			truncate(resp.Stderr, 600), truncate(resp.Stdout, 200))
	}
	assertContainsFold(t, resp.Stderr, "existing", "kill", "shutdown", "stopped")
	if resp.ExitCode != 0 {
		t.Fatalf("HandleCLI exit = %d, want 0; err=%q", resp.ExitCode, resp.KillErr)
	}
}
```