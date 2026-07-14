## Expected

Requirement **C2** (no process env overlay for session id):

- Serve healthy.
- `AgentRunCallCount == 1`.
- `AgentRunSessionID` equals live **control** session id (BuildAgentRunArgs input).
- `AgentRunSystemPath` ends with `SYSTEM.md` (case-sensitive preferred).
- Session id for agent-run is carried via **BuildAgentRunArgs** (`--session-id`
  prefixed + `--env BROWSER_AGENT_SESSION_ID=<control>`), **not** required on
  the injector env map.
- If `AgentRunEnv` contains `BROWSER_AGENT_SESSION_ID`, it must equal control id
  (optional; production must not *require* process env overlay).

## Side Effects

- Injector only; no real agent-run binary.

## Errors

- Wrong system path breaks playbook load.

## Exit Code

- Not asserted.

```go
import (
	"path/filepath"
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	if resp == nil {
		t.Fatal("resp is nil")
	}
	if resp.AgentRunCallCount != 1 {
		t.Fatalf("AgentRunFn call count=%d, want 1", resp.AgentRunCallCount)
	}
	sid := req.SessionID
	if sid == "" {
		sid = resp.RealSessionID
	}
	// First arg remains control id (disk / BuildAgentRunArgs input).
	if resp.AgentRunSessionID != sid {
		t.Fatalf("AgentRun sessionID=%q, want control %q", resp.AgentRunSessionID, sid)
	}
	sys := resp.AgentRunSystemPath
	if strings.TrimSpace(sys) == "" {
		t.Fatal("AgentRun systemPromptPath empty")
	}
	base := filepath.Base(sys)
	if base != "SYSTEM.md" && !strings.EqualFold(base, "SYSTEM.md") {
		t.Fatalf("systemPromptPath base=%q, want SYSTEM.md; full=%q", base, sys)
	}
	// Env overlay for session is optional / discouraged; if present must be control.
	if resp.AgentRunEnv != nil {
		if gotEnv, ok := resp.AgentRunEnv["BROWSER_AGENT_SESSION_ID"]; ok && gotEnv != "" {
			if gotEnv != sid {
				t.Fatalf("env BROWSER_AGENT_SESSION_ID=%q, want control %q", gotEnv, sid)
			}
		}
	}
}
```
