# Scenario

**Feature**: pure BuildAgentRunArgs argv contract

```
# no server
Test Client -> BuildAgentRunArgs(sessionID, promptPath, workspaceDir) -> []string
# must include: run, --session-id, grok-tty, auto-send-or-resume, new-terminal,
#               --no-submit (ALWAYS), --open
# --dir only when workspace non-empty
```

## Preconditions

- Mode is `agent-args`.
- No listen socket.

## Steps

1. Set `Mode = ModeAgentArgs`.
2. Children set workspace / session / prompt inputs.

## Context

- Requirement D1–D2. Env is not part of argv (tested via AgentRunFn).
- `--no-submit` is always required so serve-launched agent-run opens with a draft first prompt.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Mode = ModeAgentArgs
	if req.AgentArgsSessionID == "" {
		req.AgentArgsSessionID = "sess-agent-args"
	}
	if req.AgentArgsPromptPath == "" {
		req.AgentArgsPromptPath = "/tmp/sessions/sess-agent-args/SYSTEM.md"
	}
	return nil
}
```
