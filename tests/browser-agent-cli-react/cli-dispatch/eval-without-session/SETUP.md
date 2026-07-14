# Scenario

**Feature**: session eval without session id → dual-source error (A3)

```
HandleCLI(["session", "eval", "1+1"], env without BROWSER_AGENT_SESSION_ID)
  -> error
  -> text mentions --session-id and BROWSER_AGENT_SESSION_ID
```

## Preconditions

- No --session-id flag; env map empty / no session key.
- Nested path only (flat `eval` is unknown after complete refactor).

## Steps

1. Set CLIArgs to session eval with expression only.
2. Clear session from env.

## Context

- Must not hang waiting on network if session resolution fails first.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.CLIArgs = []string{"session", "eval", "1+1"}
	req.CLIEnv = map[string]string{}
	return nil
}
```
