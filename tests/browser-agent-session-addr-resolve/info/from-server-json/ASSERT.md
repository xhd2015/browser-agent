## Expected

After fix (Classic TDD target — **RED** on current code):

- `HandleCLI` returns nil; `ExitCode` 0.
- Stdout ends with `\n`.
- Stdout contains the live `session_id` (`resp.SessionID`).
- Stdout is JSON-ish (object with session snapshot fields).

## Side Effects

- Temp `BaseDir` and `server.json` cleaned by harness defer.

## Errors

- `session not found`, `status 404`, or empty stdout fails (current bug symptom).

## Exit Code

- 0.

```go
import (
	"encoding/json"
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	if resp == nil {
		t.Fatal("resp is nil")
	}
	if resp.DispatchTimedOut {
		t.Fatal("session info timed out")
	}
	if !resp.DaemonMetaExists {
		t.Fatal("server.json must exist for from-server-json leaf")
	}
	if resp.SessionID == "" {
		t.Fatal("harness did not create session")
	}

	if resp.CLIErr != "" {
		t.Fatalf("session info without --addr should succeed when server.json points at live daemon; CLIErr=%q stderr=%q stdout=%q",
			resp.CLIErr, resp.Stderr, resp.Stdout)
	}
	assertExitZero(t, resp)
	assertStdoutTrailingNewline(t, resp.Stdout)

	if !strings.Contains(resp.Stdout, resp.SessionID) {
		t.Fatalf("stdout missing session_id %q; stdout=%s", resp.SessionID, truncate(resp.Stdout, 500))
	}

	trim := strings.TrimSpace(resp.Stdout)
	var m map[string]any
	if err := json.Unmarshal([]byte(trim), &m); err != nil {
		t.Fatalf("stdout should be JSON object; parse err=%v stdout=%s", err, truncate(resp.Stdout, 400))
	}
	if sid, ok := m["session_id"].(string); !ok || sid != resp.SessionID {
		t.Fatalf("JSON session_id=%v want %q; stdout=%s", m["session_id"], resp.SessionID, truncate(resp.Stdout, 400))
	}
}
```