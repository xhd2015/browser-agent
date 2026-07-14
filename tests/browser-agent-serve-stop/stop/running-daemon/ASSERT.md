## Expected

- First daemon health goes down after `serve --stop`.
- `{BaseDir}/server.json` is removed.
- `HandleCLI` returns **nil** quickly (no second bind / no blocking serve).
- Stderr contains operator-facing stop/shutdown/kill status (case-insensitive).

## Side Effects

- Removes running daemon `server.json`; does not leave a replacement daemon.

## Errors

- Daemon still healthy, meta still present, missing stderr status, or non-nil CLI error fails.

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
	if resp.ExitCode != 0 {
		t.Fatalf("HandleCLI exit = %d, want 0; cliErr=%q stderr=%q",
			resp.ExitCode, resp.CLIErr, truncate(resp.Stderr, 400))
	}
	if !resp.HealthDown {
		t.Fatal("daemon health still up after serve --stop")
	}
	if resp.DaemonMetaExists {
		t.Fatalf("server.json still exists after serve --stop; path=%s", resp.DaemonMetaPath)
	}
	if !resp.StopMessageSeen {
		t.Fatalf("stderr missing stop status; stderr=%q stdout=%q",
			truncate(resp.Stderr, 600), truncate(resp.Stdout, 200))
	}
	assertContainsFold(t, resp.Stderr, "stop", "stopped", "shutdown", "kill")
}
```