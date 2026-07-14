## Expected

- `HandleCLI session new` returns **nil** (`CLIErr` empty).
- Session `sess-cli-8` registered on `GET /v1/sessions`.
- Stdout contains session id and `browser-agent session info`.
- `OpenChromeCallCount == 1` via `SessionNewTestHooks`.

## Side Effects

- Daemon ensured/spawned; session created; Chrome hook recorded only.

## Errors

- Non-nil CLI error, missing session, or zero OpenChrome calls fails.

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
	if resp.CLIErr != "" {
		t.Fatalf("HandleCLI error: %s", resp.CLIErr)
	}
	if !resp.SessionOnServer {
		t.Fatalf("session %q not in GET /v1/sessions: %v", req.SessionID, resp.ServerSessionIDs)
	}
	if resp.OpenChromeCallCount != 1 {
		t.Fatalf("OpenChromeFn call count=%d, want 1", resp.OpenChromeCallCount)
	}
	assertContainsFold(t, resp.Stdout, req.SessionID, "browser-agent session info")
}```
