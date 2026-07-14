# Scenario

**Feature**: list shows multiple sessions sorted by id

```
POST sess-alpha, POST sess-zulu -> GET /v1/sessions sorted
HandleCLI session list --base-dir BaseDir
  -> both ids in stdout; alpha before zulu
```

## Preconditions

- Two explicit session ids for deterministic sort order.

## Steps

1. Set `ListOp = ListOpTwoSessions`.
2. `SessionIDsToCreate = ["sess-alpha", "sess-zulu"]`.

## Context

- API returns snapshots sorted by `session_id`; human table should follow same order.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.ListOp = ListOpTwoSessions
	req.SessionIDsToCreate = []string{"sess-alpha", "sess-zulu"}
	req.JSONMode = false
	return nil
}```
