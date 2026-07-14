# Scenario

**Feature**: HandleCLI `session list` against live daemon or daemon-down path

```
RunDaemon -> POST /v1/sessions (optional)
HandleCLI session list --base-dir BaseDir [--json]
  -> table or JSON array

no daemon -> warning stderr; exit 0
```

## Preconditions

- Mode `list`; leaf Setup sets `ListOp` and session seeds.

## Steps

1. Set `Mode = ModeList`.
2. Leaf Setup sets `ListOp`, `SessionIDsToCreate`, `JSONMode`, and daemon flags.

## Context

- Omit `--addr` by default; addr resolves from `server.json` (session-addr-resolve pattern).

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Mode = ModeList
	return nil
}```
