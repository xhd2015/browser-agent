# Scenario

**Feature**: `session list` human table with wider columns and footer hints

```
RunDaemon -> create session(s) [optional fake hello telemetry]
HandleCLI session list --base-dir BaseDir
```

## Preconditions

- Mode is `list`.
- Human default output (no `--json`).

## Steps

1. Set `Mode = ModeList`.
2. Leaves set `ListOp`.

## Context

- Columns: Session ID, Created, Pages, Browser, Status.
- Unknown pages → `—`; 0-page footer hints `session delete`.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Mode = ModeList
	return nil
}
```