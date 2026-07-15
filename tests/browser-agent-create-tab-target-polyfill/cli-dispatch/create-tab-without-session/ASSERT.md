## Expected

Requirement **A2**:

- Non-nil HandleCLI error (CLIErr set).
- Text mentions `--session-id` and `BROWSER_AGENT_SESSION_ID`.
- DispatchTimedOut false.

## Side Effects

- None.

## Errors

- Silent success is a failure.

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
		t.Fatal("create-tab without session timed out")
	}
	if resp.CLIErr == "" {
		t.Fatal("expected HandleCLI error when create-tab has no session")
	}
	text := resp.CLIErr + "\n" + combinedCLIText(resp)
	assertSessionSourcesInText(t, text)
}
```
