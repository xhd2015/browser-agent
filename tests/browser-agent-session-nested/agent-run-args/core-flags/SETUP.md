# Scenario

**Feature**: BuildAgentRunArgs core flags — prefix, --env, --no-submit (A3)

```
BuildAgentRunArgs("ctrl-core-1", "/tmp/sessions/ctrl-core-1/SYSTEM.md", "/work/ws")
  -> run
  -> --session-id=browser-agent-sess-ctrl-core-1
  -> --env BROWSER_AGENT_SESSION_ID=ctrl-core-1
  -> --no-submit
  -> grok-tty, auto-send-or-resume, new-terminal, --dir, --open
```

## Preconditions

- Non-empty workspace so --dir may appear (not required for A3 core).
- Control id `ctrl-core-1` is unique and unprefixed.

## Steps

1. Set AgentArgsControlID, prompt path, workspace.

## Context

- Requirement A3.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.AgentArgsControlID = "ctrl-core-1"
	req.AgentArgsPromptPath = "/tmp/sessions/ctrl-core-1/SYSTEM.md"
	req.AgentArgsWorkspace = "/work/project-api-capture"
	return nil
}
```
