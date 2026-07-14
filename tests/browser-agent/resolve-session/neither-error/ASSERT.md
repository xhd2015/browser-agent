## Expected

Requirement **A3** (and CLI missing-session path **C4**):

- Resolve returns an error (non-empty `ResolveErr`).
- Error message mentions `--session-id`.
- Error message mentions `BROWSER_AGENT_SESSION_ID`.
- ResolvedID is empty.

## Side Effects

- None.

## Errors

- Silent empty id without error is a failure.
- Error that only mentions one of flag/env is insufficient.

## Exit Code

- Not asserted (package API).

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
	if resp.ResolveErr == "" {
		t.Fatal("expected resolve error when neither flag nor env set")
	}
	if resp.ResolvedID != "" {
		t.Fatalf("ResolvedID = %q, want empty on error", resp.ResolvedID)
	}
	msg := resp.ResolveErr
	if !strings.Contains(msg, "--session-id") {
		t.Fatalf("error must mention --session-id; got %q", msg)
	}
	if !strings.Contains(msg, "BROWSER_AGENT_SESSION_ID") {
		t.Fatalf("error must mention BROWSER_AGENT_SESSION_ID; got %q", msg)
	}
}
```
