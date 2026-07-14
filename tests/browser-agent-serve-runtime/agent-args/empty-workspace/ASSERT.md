## Expected

Requirement **D2**:

- Core argv tokens still present (`run`, prefixed `--session-id`, `--env BROWSER_AGENT_SESSION_ID`, grok-tty, auto-send-or-resume, new-terminal, **`--no-submit`**, open).
- **No** `--dir` flag (empty workspace omitted).
- **`--no-submit` always present** even when workspace is empty (draft open is workspace-independent).

## Side Effects

- None (pure).

## Errors

- Emitting empty `--dir` confuses agent-run working directory.
- Omitting `--no-submit` would auto-submit the first prompt in TTY.

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
	assertAgentRunArgsCore(t, resp.AgentRunArgs, req.AgentArgsSessionID, false)
	// Explicit leaf-level require for --no-submit (always, even with empty workspace)
	if !argvHasToken(resp.AgentRunArgs, "--no-submit") {
		found := false
		for _, a := range resp.AgentRunArgs {
			if a == "--no-submit" || strings.HasPrefix(a, "--no-submit=") {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("D2: argv must include --no-submit (workspace-independent); args=%v", resp.AgentRunArgs)
		}
	}
}
```
