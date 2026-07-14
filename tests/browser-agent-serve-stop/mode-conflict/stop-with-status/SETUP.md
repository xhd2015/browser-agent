# Scenario

**Feature**: --stop and --status together are rejected

```
HandleCLI serve --stop --status -> mutually exclusive error -> exit 1
```

## Preconditions

- No daemon required; parse-time rejection.

## Steps

1. Set `ModeConflictOp = ModeConflictOpStopStatus`.

## Context

- `--status` is read-only probe; must not combine with `--stop`.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.ModeConflictOp = ModeConflictOpStopStatus
	return nil
}
```