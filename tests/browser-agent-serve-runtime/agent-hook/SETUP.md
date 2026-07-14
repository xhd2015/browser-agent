# Scenario

**Feature**: AgentRunFn honors NoAgentRun (injectable; never real agent-run)

```
# skipped
NoAgentRun=true + AgentRunFn set -> fn never called

# called
NoAgentRun=false + AgentRunFn records
  -> once(sessionID, systemPromptPath ends SYSTEM.md, env BROWSER_AGENT_SESSION_ID)
```

## Preconditions

- Mode is `agent-hook`.
- Chrome isolated off (`NoOpenChrome=true`).
- Harness injects recording AgentRunFn.

## Steps

1. Set `Mode = ModeAgentHook`.
2. Children set `HookExpect`.

## Context

- Requirement C1–C2.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Mode = ModeAgentHook
	req.NoOpenChrome = true
	req.InjectAgentRunFn = true
	return nil
}
```
