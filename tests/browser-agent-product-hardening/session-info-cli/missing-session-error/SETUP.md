# Scenario

**Feature**: session info without session id → dual-source error (B3)

```
HandleCLI(["session", "info"], empty env)
  -> non-nil error
  -> mentions --session-id and BROWSER_AGENT_SESSION_ID
```

## Preconditions

- No --session-id; empty CLIEnv (no ambient process env fallback).
- No live server required.

## Steps

1. Set SessionInfoKind missing-session-error.
2. CLIArgs default to `session info` in Run if empty.

## Context

- Requirement B3. Same resolve contract as other nested side-commands.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.SessionInfoKind = SessionInfoMissingSession
	req.CLIArgs = nil
	req.CLIEnv = map[string]string{}
	return nil
}
```
