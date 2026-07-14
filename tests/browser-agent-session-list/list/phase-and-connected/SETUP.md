# Scenario

**Feature**: list shows Phase and Connected columns for mixed session states

```
POST sess-wait-list (no WS) -> waiting_extension
POST sess-conn-list + fake WS hello -> extension_connected
HandleCLI session list --base-dir BaseDir
  -> phase/connected columns reflect each session
```

## Preconditions

- One waiting session; one connected via fake extension (phase4 harness).

## Steps

1. Set `ListOp = ListOpPhaseConnected`.
2. `SessionIDsToCreate = ["sess-wait-list", "sess-conn-list"]`.
3. `ConnectExtensionFor = "sess-conn-list"`.

## Context

- Human columns: Phase (`waiting_extension` / `extension_connected`), Connected (`no` / `yes`).

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.ListOp = ListOpPhaseConnected
	req.SessionIDsToCreate = []string{"sess-wait-list", "sess-conn-list"}
	req.ConnectExtensionFor = "sess-conn-list"
	req.JSONMode = false
	return nil
}```
