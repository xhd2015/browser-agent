## Expected

After implementer lands session list (**RED** on current code):

- `HandleCLI` returns nil; `ExitCode` 0.
- Stdout ends with `\n`.
- Stdout contains `0 sessions` or table header `Session ID` with no session rows before summary.
- Stdout does not contain `unknown session subcommand`.

## Side Effects

- Read-only list; no session mutation.

## Errors

- Non-zero exit, missing zero-session indication, or unknown subcommand error fails.

## Exit Code

- 0.

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
	if resp.DispatchTimedOut {
		t.Fatal("session list timed out")
	}
	if !resp.DaemonMetaExists {
		t.Fatal("server.json must exist for empty list leaf")
	}

	assertExitZero(t, resp)
	assertStdoutTrailingNewline(t, resp.Stdout)

	low := strings.ToLower(resp.Stdout)
	if strings.Contains(low, "unknown session subcommand") {
		t.Fatalf("session list not implemented; stdout=%q stderr=%q cliErr=%q",
			truncate(resp.Stdout, 400), truncate(resp.Stderr, 200), resp.CLIErr)
	}

	hasZero := strings.Contains(low, "0 sessions") ||
		(strings.Contains(low, "session id") && strings.Contains(low, "0 session"))
	if !hasZero {
		// Accept empty table: header present, no sess- ids, summary mentions 0
		hasHeader := strings.Contains(low, "session id")
		hasSessID := strings.Contains(low, "sess-")
		if !(hasHeader && !hasSessID) {
			t.Fatalf("stdout missing zero-session indication; stdout=%q", truncate(resp.Stdout, 600))
		}
	}

	if len(resp.APISessionIDs) != 0 {
		t.Fatalf("API reported %d sessions want 0: %v", len(resp.APISessionIDs), resp.APISessionIDs)
	}
}```
