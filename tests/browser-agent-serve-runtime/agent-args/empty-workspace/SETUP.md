# Scenario

**Feature**: BuildAgentRunArgs with empty workspace omits --dir; still prefix + --env (D2)

```
BuildAgentRunArgs(id, SYSTEM.md, "")
  -> no --dir flag
  -> --session-id=browser-agent-sess-<id>
  -> --env BROWSER_AGENT_SESSION_ID=<id>
  -> still --no-submit
```

## Preconditions

- AgentArgsWorkspace empty string.

## Steps

1. Set AgentArgsWorkspace to `""`.

## Context

- Empty dir must not produce `--dir=` or `--dir ""`.
- `--no-submit` is still always present (workspace-independent).

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.AgentArgsWorkspace = ""
	req.AgentArgsSessionID = "sess-empty-ws"
	req.AgentArgsPromptPath = "/tmp/sessions/sess-empty-ws/SYSTEM.md"
	return nil
}
```
