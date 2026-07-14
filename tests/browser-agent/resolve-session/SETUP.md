# Scenario

**Feature**: resolve browser-agent session id from flag vs env

```
# CLI / package helper — no Control Server
Operator supplies --session-id and/or BROWSER_AGENT_SESSION_ID
Test Client -> ResolveSessionID(flag, env)
  -> session id (flag wins) | hard error naming both sources
```

## Preconditions

- Mode is pure resolve (`ModeResolveSession`).
- No HTTP listen required.
- Flag wins when set; else env; else error.

## Steps

1. Set `Mode = ModeResolveSession` (`"resolve-session"`).
2. Descendants set FlagSet/FlagValue and EnvSet/EnvValue.

## Context

- Covers requirement A1–A3 and C4 (missing session on CLI path).

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Mode = ModeResolveSession
	return nil
}
```
