# Scenario

**Feature**: pure BuildAgentRunArgs with prefixed session + --env + --no-submit

```
Test Client -> BuildAgentRunArgs(control, SYSTEM.md, workspace) -> []string
# --session-id=<agent-run-id>
# --env BROWSER_AGENT_SESSION_ID=<control>
# --no-submit always
```

## Preconditions

- Mode is agent-run-args.
- Control id is distinct from agent-run id (unless already prefixed).

## Steps

1. Set Mode = ModeAgentRunArgs.
2. Leave control / prompt / workspace to leaf.

## Context

- launchAgentRun must not rely on process env overlay for session id.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Mode = ModeAgentRunArgs
	return nil
}
```
