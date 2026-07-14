# Scenario

**Feature**: session info without session id → dual-source error (A4)

```
HandleCLI(["session", "info"], env without BROWSER_AGENT_SESSION_ID)
  -> error
  -> text mentions --session-id and BROWSER_AGENT_SESSION_ID
```

## Preconditions

- No --session-id; empty env map.
- Nested path only.

## Steps

1. Set CLIArgs to `session info` only.
2. Clear session from env.

## Context

- Same session resolution path as session eval.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.CLIArgs = []string{"session", "info"}
	req.CLIEnv = map[string]string{}
	return nil
}
```
