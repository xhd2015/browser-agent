# Scenario

**Feature**: session list --json emits raw JSON array

```
RunDaemon -> POST sess-json-mode
HandleCLI session list --json --base-dir BaseDir
  -> valid JSON array; no human table prose
```

## Preconditions

- One session seeded for non-trivial JSON payload.

## Steps

1. Set `ListOp = ListOpJSONMode`.
2. `SessionIDsToCreate = ["sess-json-mode"]`.
3. `JSONMode = true`.

## Context

- `--json` must not emit ANSI or table headers (`Session ID` column prose).

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.ListOp = ListOpJSONMode
	req.SessionIDsToCreate = []string{"sess-json-mode"}
	req.JSONMode = true
	return nil
}```
