# Scenario

**Feature**: session create-tab without session fails with dual sources (A2)

```
HandleCLI(["session", "create-tab"], empty env)
  -> non-nil error
  -> mentions --session-id and BROWSER_AGENT_SESSION_ID
```

## Preconditions

- DispatchKind = create-tab-without-session.
- CLIEnv empty (no ambient session).

## Steps

1. Set DispatchCreateTabWithoutSession.
2. Leave CLIArgs empty so Run injects `session create-tab`.

## Context

- Requirement A2.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.DispatchKind = DispatchCreateTabWithoutSession
	req.CLIArgs = nil
	req.CLIEnv = map[string]string{}
	return nil
}
```
