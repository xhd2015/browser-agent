## Expected

Requirement **A3**:

- Non-nil HandleCLI error (CLIErr set).
- Combined CLIErr + stdout + stderr mentions `--session-id`.
- Same text mentions `BROWSER_AGENT_SESSION_ID`.
- DispatchTimedOut false.

## Side Effects

- None (no job POST without session).

## Errors

- Silent success or single-source mention is a failure.

## Exit Code

- Non-zero.

```go
import (
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	if resp == nil {
		t.Fatal("resp is nil")
	}
	if resp.DispatchTimedOut {
		t.Fatal("session eval without session timed out")
	}
	if resp.CLIErr == "" {
		t.Fatal("expected HandleCLI error when session eval has no session")
	}
	text := resp.CLIErr + "\n" + combinedCLIText(resp)
	assertSessionSourcesInText(t, text)
}
```
