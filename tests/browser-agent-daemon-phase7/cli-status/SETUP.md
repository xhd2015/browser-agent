# Scenario

**Feature**: `HandleCLI serve --status` read-only operator command

```
HandleCLI serve --status --base-dir <dir> -> stdout table -> exit 0
```

## Preconditions

- Mode `ModeCLIStatus`.
- Leaf sets `CLIStatusOp`.

## Steps

1. Set `Mode = ModeCLIStatus`.

## Context

- Must not call `RunDaemon` or spawn Chrome.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Mode = ModeCLIStatus
	return nil
}
```