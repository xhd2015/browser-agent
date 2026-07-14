# Scenario

**Feature**: cdp without session id fails with dual sources (A5)

```
HandleCLI(["session", "cdp", "Page.navigate", json], empty env)
  -> non-nil error
  -> mentions --session-id and BROWSER_AGENT_SESSION_ID
```

## Preconditions

- DispatchKind = cdp-without-session.

## Steps

1. Set DispatchKind to DispatchCDPWithoutSession.

## Context

- Requirement A5.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.DispatchKind = DispatchCDPWithoutSession
	req.CLIArgs = nil
	return nil
}
```
