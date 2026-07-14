# Scenario

**Feature**: warn disconnected orphans

```
serve --kill-existing -> warn connected + RemoveAll orphan dirs
```

## Preconditions

- `KillExistingOp = KillExistingOpWarnOrphans`.

## Steps

1. Set `KillExistingOp = KillExistingOpWarnOrphans`.
2. Set leaf-specific request fields.

## Context

- See ASSERT.md for expected outcomes.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.KillExistingOp = KillExistingOpWarnOrphans
	req.OrphanID = "sess-orph02"
	return nil
}
```
