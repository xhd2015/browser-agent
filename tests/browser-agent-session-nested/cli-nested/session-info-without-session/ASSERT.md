## Expected

Requirement **C2**:

- Non-nil HandleCLI error (CLIErr non-empty).
- Combined error/printed text mentions `--session-id` and `BROWSER_AGENT_SESSION_ID`.
- `DispatchTimedOut` false.

## Side Effects

- Must not hang on network when session resolve fails first.

## Errors

- Silent success without session is a failure.

## Exit Code

- Non-zero preferred.

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
		t.Fatal("HandleCLI timed out on session info without session")
	}
	if resp.CLIErr == "" {
		t.Fatal("expected non-nil HandleCLI error for session info without session")
	}
	assertSessionResolveErrorText(t, combinedCLIText(resp))
}
```
