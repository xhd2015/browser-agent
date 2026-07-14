## Expected

Fallback path (typically **GREEN** before and after fix):

- `server.json` absent under `BaseDir`.
- CLI fails (non-zero / CLIErr) because no daemon at default addr **or** unknown
  session on a stale `:43761` daemon — both acceptable.
- Combined output must **not** reference the ephemeral harness addr (proves meta was
  not read from a non-existent file).
- Must **not** succeed with session JSON for a session that was never created in
  this test.

## Side Effects

- None.

## Errors

- Unexpected success with fabricated session snapshot fails.

## Exit Code

- Non-zero expected; not strictly asserted.

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
		t.Fatal("session info timed out in harness")
	}
	if resp.DaemonMetaExists {
		t.Fatal("server.json must be absent for fallback leaf")
	}

	combined := combinedOutput(resp)
	low := strings.ToLower(combined)

	// If CLI succeeded, it must not be via ephemeral meta (no daemon was started).
	if resp.ExitCode == 0 && resp.CLIErr == "" {
		if strings.Contains(resp.Stdout, "sess-fallback-no-meta") {
			// Only OK if this is a genuine 43761 response for unknown id — still suspicious.
			t.Fatalf("unexpected success without server.json; stdout=%s", truncate(resp.Stdout, 400))
		}
	}

	// Accept connection / request errors toward default addr.
	connOK := strings.Contains(low, "connection refused") ||
		strings.Contains(low, "connect:") ||
		strings.Contains(low, "request failed") ||
		strings.Contains(low, "dial tcp") ||
		strings.Contains(low, "status 404") ||
		strings.Contains(low, "session not found") ||
		resp.CLIErr != ""
	if !connOK && resp.ExitCode == 0 {
		t.Fatalf("expected failure or connection error without meta; stdout=%q stderr=%q CLIErr=%q",
			resp.Stdout, resp.Stderr, resp.CLIErr)
	}
}
```