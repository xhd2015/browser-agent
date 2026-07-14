# Scenario

**Feature**: flag wins when both flag and env are set (A1)

```
ResolveSessionID(flag=sess-flag, env=sess-env) -> sess-flag
```

## Preconditions

- FlagSet=true with FlagValue `sess-from-flag`.
- EnvSet=true with EnvValue `sess-from-env` (different).

## Steps

1. Set flag value `sess-from-flag` (FlagSet true).
2. Set env value `sess-from-env` (EnvSet true).

## Context

- Env must not shadow an explicit flag.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.FlagSet = true
	req.FlagValue = "sess-from-flag"
	req.EnvSet = true
	req.EnvValue = "sess-from-env"
	return nil
}
```
