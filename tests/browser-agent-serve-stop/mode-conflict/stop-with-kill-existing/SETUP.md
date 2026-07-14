# Scenario

**Feature**: --stop and --kill-existing together are rejected

```
HandleCLI serve --stop --kill-existing -> mutually exclusive error -> exit 1
```

## Preconditions

- No daemon required; parse-time rejection.

## Steps

1. Set `ModeConflictOp = ModeConflictOpStopKillExisting`.

## Context

- Both flags invoke kill paths; only one mode flag may be set.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.ModeConflictOp = ModeConflictOpStopKillExisting
	return nil
}
```