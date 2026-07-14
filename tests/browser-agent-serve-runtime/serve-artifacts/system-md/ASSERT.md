## Expected

Requirement **A1** (nested session playbook):

- Serve succeeds (no transport error).
- `SYSTEM.md` exists under `{BaseDir}/sessions/{sessionID}/`.
- Text does **not** embed the live control session id.
- Text contains nested CLI recipes:
  - `browser-agent session info`
  - `browser-agent session eval`
  - `browser-agent session run`
  - `browser-agent session logs`
  - `browser-agent session screenshot`
- Mentions `BROWSER_AGENT_SESSION_ID`.

## Side Effects

- Session dir created under BaseDir only.
- No Chrome / agent-run processes.

## Errors

- Missing file or incomplete nested recipes fails agent playbook bootstrap.
- Embedding control id fails open-prompt isolation.

## Exit Code

- Not asserted (in-process serve).

```go
import (
	"os"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	if resp == nil {
		t.Fatal("resp is nil")
	}
	sid := req.SessionID
	if sid == "" {
		sid = resp.RealSessionID
	}
	if resp.SystemMDPath == "" {
		t.Fatal("SystemMDPath empty")
	}
	if _, err := os.Stat(resp.SystemMDPath); err != nil {
		t.Fatalf("SYSTEM.md missing at %s: %v", resp.SystemMDPath, err)
	}
	assertSystemMDRecipes(t, resp.SystemMDText, sid)
}
```
