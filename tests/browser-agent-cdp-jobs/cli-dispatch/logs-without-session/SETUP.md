# Scenario

**Feature**: logs without session id fails with dual sources (A3)

```
HandleCLI(["session", "logs"], empty env)
  -> non-nil error
  -> mentions --session-id and BROWSER_AGENT_SESSION_ID
```

## Preconditions

- DispatchKind = logs-without-session.

## Steps

1. Set DispatchKind to DispatchLogsWithoutSession.

## Context

- Requirement A3.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.DispatchKind = DispatchLogsWithoutSession
	req.CLIArgs = nil
	return nil
}
```
