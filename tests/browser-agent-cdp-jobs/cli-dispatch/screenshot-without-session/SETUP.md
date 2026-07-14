# Scenario

**Feature**: screenshot without session id fails with dual sources (A4)

```
HandleCLI(["session", "screenshot"], empty env)
  -> non-nil error
  -> mentions --session-id and BROWSER_AGENT_SESSION_ID
```

## Preconditions

- DispatchKind = screenshot-without-session.

## Steps

1. Set DispatchKind to DispatchShotWithoutSession.

## Context

- Requirement A4.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.DispatchKind = DispatchShotWithoutSession
	req.CLIArgs = nil
	return nil
}
```
