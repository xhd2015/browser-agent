## Expected

After implementer lands session-rich (**RED** on current code):

- `HandleCLI` returns nil; `ExitCode` 0.
- Stdout is human-formatted (not JSON-only).
- Stdout contains session id.
- Stdout shows Status **ready** (label or human line).
- Stdout mentions Pages count `1` or single-page context.

## Side Effects

- Fake extension stays connected during info.

## Errors

- JSON-only output or missing ready status fails.

## Exit Code

- 0.

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
		t.Fatal("session info timed out")
	}

	assertExitZero(t, resp)
	assertStdoutTrailingNewline(t, resp.Stdout)

	if stdoutLooksLikeJSONObject(resp.Stdout) {
		t.Fatalf("default session info must be human output; stdout=%s", truncate(resp.Stdout, 500))
	}

	assertContainsFold(t, resp.Stdout, resp.SessionID, "status", "ready")
	assertContainsFold(t, resp.Stdout, "pages", "1")
}
```