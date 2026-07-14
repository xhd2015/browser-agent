# Scenario

**Feature**: NoAgentRun=false → AgentRunFn called once with control session + SYSTEM.md (C2)

```
Config{NoAgentRun:false, AgentRunFn:record, WorkspaceDir}
  -> once; system path ends SYSTEM.md
  -> session for agent-run via BuildAgentRunArgs (--env / prefixed --session-id)
  -> process env overlay for session id not required
```

## Preconditions

- HookExpect = called.
- NoAgentRun false; NoOpenChrome true.
- WorkspaceDir prepared by Run when empty.

## Steps

1. Set HookExpect called.
2. Set NoAgentRun false.

## Context

- launchAgentRun must not depend on manual cmd.Env overlay for session id.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.HookExpect = HookExpectCalled
	req.NoAgentRun = false
	return nil
}
```
