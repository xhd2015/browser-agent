## Expected

Regression guard (may be **GREEN** before fix):

- `HandleCLI` nil error; `ExitCode` 0.
- Stdout ends with `\n`.
- Stdout contains live `session_id`.

## Side Effects

- None beyond temp dirs.

## Errors

- CLIErr or missing session id fails.

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
		t.Fatal("session info timed out")
	}
	if resp.CLIErr != "" {
		t.Fatalf("explicit --addr should reach live daemon; CLIErr=%q stderr=%q stdout=%q",
			resp.CLIErr, resp.Stderr, resp.Stdout)
	}
	assertExitZero(t, resp)
	assertStdoutTrailingNewline(t, resp.Stdout)

	if resp.SessionID == "" {
		t.Fatal("harness did not create session")
	}
	if !strings.Contains(resp.Stdout, resp.SessionID) {
		t.Fatalf("stdout missing session_id %q; stdout=%s", resp.SessionID, truncate(resp.Stdout, 500))
	}
}
```