# Scenario

**Feature**: pure AgentRunSessionID prefix mapping

```
Test Client -> AgentRunSessionID(controlID) -> agent-run id
# bare control gets prefix; already-prefixed is idempotent
```

## Preconditions

- Mode is agent-run-id.
- Package exports AgentRunSessionID (RED until implementer lands it).

## Steps

1. Set Mode = ModeAgentRunID.
2. Leave ControlSessionID to leaf Setup.

## Context

- Prefix literal: browser-agent-sess-

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Mode = ModeAgentRunID
	return nil
}
```
