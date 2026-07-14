# Scenario

**Feature**: BuildAgentRunArgs with workspace includes --dir + prefix + --env (D1)

```
BuildAgentRunArgs(id, SYSTEM.md, "/work/project")
  -> run, --session-id=browser-agent-sess-<id>, --env BROWSER_AGENT_SESSION_ID=<id>
  -> grok-tty, auto-send-or-resume, new-terminal, --dir, --no-submit, --open
```

## Preconditions

- Non-empty AgentArgsWorkspace.

## Steps

1. Set AgentArgsWorkspace to a non-empty absolute-style path.

## Context

- Dir flag form may be `--dir=path` or `--dir path`.
- Control id remains `sess-ws-1`; agent-run id is prefixed.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.AgentArgsWorkspace = "/work/project-api-capture"
	req.AgentArgsSessionID = "sess-ws-1"
	req.AgentArgsPromptPath = "/tmp/sessions/sess-ws-1/SYSTEM.md"
	return nil
}
```
