# Scenario

**Feature**: SessionRegistry List — sorted session snapshots

```
Test Client -> List() -> []sessionSnapshot sorted by session_id
```

## Preconditions

- Mode is list.
- Leaf sets ListSessionIDs (empty for empty registry).

## Steps

1. Set Mode to list.
2. Ensure BaseDir and Addr.

## Context

- Stable ascending sort by session id string.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Mode = ModeList
	ensureBaseDir(t, req)
	ensureAddr(t, req)
	return nil
}
```