# Scenario

**Feature**: KillExistingDaemon client helper

```
RunDaemon + server.json
KillExistingDaemon(baseDir, timeout) -> POST shutdown -> wait -> (force kill) -> meta gone
```

## Preconditions

- Mode `ModeKillExisting`.
- Leaf sets `KillExistingOp`.

## Steps

1. Set `Mode = ModeKillExisting`.

## Context

- Reads `{BaseDir}/server.json` for base_url and pid.
- Default kill timeout 10s unless leaf overrides.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Mode = ModeKillExisting
	return nil
}
```