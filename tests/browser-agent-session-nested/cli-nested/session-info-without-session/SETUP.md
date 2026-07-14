# Scenario

**Feature**: session info without session id → dual-source error (C2)

```
HandleCLI(["session", "info"], empty env)
  -> non-nil error
  -> mentions --session-id and BROWSER_AGENT_SESSION_ID
```

## Preconditions

- No --session-id; empty CLIEnv.

## Steps

1. Set CLIKind session-info-without-session.

## Context

- Requirement C2. Nested path still uses shared resolve.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.CLIKind = CLIKindSessionInfoNoSession
	req.CLIArgs = nil
	req.CLIEnv = map[string]string{}
	return nil
}
```
