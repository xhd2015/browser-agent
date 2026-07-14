# Scenario

**Feature**: env alone supplies session id (A2)

```
ResolveSessionID(flag unset, env=sess-from-env) -> sess-from-env
```

## Preconditions

- FlagSet=false.
- EnvSet=true with EnvValue `sess-from-env`.

## Steps

1. Clear flag (FlagSet false).
2. Set env `sess-from-env`.

## Context

- Typical agent-run injection of `BROWSER_AGENT_SESSION_ID`.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.FlagSet = false
	req.FlagValue = ""
	req.EnvSet = true
	req.EnvValue = "sess-from-env"
	return nil
}
```
