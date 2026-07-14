## Expected

- `SessionNew` returns **nil** error.
- `OpenChromeCallCount == 0`.
- `AgentRunProbeCallCount == 0`.
- Session id `sess-new-8` appears in `GET /v1/sessions`.
- Stdout contains `session-id:` and `Session URL:` markers.

## Side Effects

- Session registered on daemon; Chrome hook not invoked.

## Errors

- OpenChrome called, agent-run probe invoked, missing stdout markers, or session not on server fails.

## Exit Code

- Not asserted.

```go
import (
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	if resp == nil {
		t.Fatal("resp is nil")
	}
	if resp.SessionNewErr != "" {
		t.Fatalf("SessionNew error: %s", resp.SessionNewErr)
	}
	if resp.OpenChromeCallCount != 0 {
		t.Fatalf("OpenChromeFn call count=%d, want 0 when NoOpenChrome", resp.OpenChromeCallCount)
	}
	if resp.AgentRunProbeCallCount != 0 {
		t.Fatalf("AgentRunProbeFn called %d times; want 0 (no agent-run)", resp.AgentRunProbeCallCount)
	}
	sid := req.SessionID
	if !resp.SessionOnServer {
		t.Fatalf("session %q not in GET /v1/sessions: %v", sid, resp.ServerSessionIDs)
	}
	if resp.Stdout == "" {
		t.Fatal("stdout is empty")
	}
	assertContainsFold(t, resp.Stdout, "session-id:", "session url:")
}
```