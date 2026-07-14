# Scenario

**Feature**: run without session id fails with dual sources (A2)

```
HandleCLI(["session", "run", "script.js"], empty env)
  -> non-nil error
  -> mentions --session-id and BROWSER_AGENT_SESSION_ID
```

## Preconditions

- DispatchKind = run-without-session.
- No ambient session env.

## Steps

1. Set DispatchKind to DispatchRunWithoutSession.

## Context

- Requirement A2. File need not exist if session check runs first.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.DispatchKind = DispatchRunWithoutSession
	req.CLIArgs = nil
	return nil
}
```
