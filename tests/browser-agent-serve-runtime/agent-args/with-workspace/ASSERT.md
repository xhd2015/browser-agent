## Expected

Requirement **D1**:

- Non-empty argv.
- Contains:
  - `run`
  - `--session-id=browser-agent-sess-<control>` (prefixed agent-run id)
  - `--env BROWSER_AGENT_SESSION_ID=<control>`
  - `grok-tty` (via `--agent-runner=grok-tty` or equivalent)
  - `--auto-send-or-resume`
  - `--new-terminal`
  - `--dir` with non-empty value (workspace)
  - **`--no-submit`** (always — draft first prompt; no auto-submit)
  - `--open`
- Prompt / SYSTEM path may appear after `--` (optional assert: path fragment present soft).

## Side Effects

- None (pure).

## Errors

- Missing required flags (including `--no-submit`) breaks agent-run launch parity with production CLI.

## Exit Code

- Not asserted.

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
	assertAgentRunArgsCore(t, resp.AgentRunArgs, req.AgentArgsSessionID, true)
	// Prefer workspace path appears near --dir
	joined := strings.Join(resp.AgentRunArgs, " ")
	if req.AgentArgsWorkspace != "" && !strings.Contains(joined, req.AgentArgsWorkspace) {
		t.Fatalf("argv should include workspace %q; args=%v", req.AgentArgsWorkspace, resp.AgentRunArgs)
	}
	// Explicit leaf-level require for --no-submit (always)
	if !argvHasToken(resp.AgentRunArgs, "--no-submit") {
		found := false
		for _, a := range resp.AgentRunArgs {
			if a == "--no-submit" || strings.HasPrefix(a, "--no-submit=") {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("D1: argv must include --no-submit; args=%v", resp.AgentRunArgs)
		}
	}
}
```
