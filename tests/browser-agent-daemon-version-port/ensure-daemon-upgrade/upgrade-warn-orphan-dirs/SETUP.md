# Scenario

**Feature**: upgrade warn orphan dirs

```
EnsureDaemon + CompareVersion -> reuse|warn|kill+respawn
```

## Preconditions

- `UpgradeOp = UpgradeOpUpgradeWarnOrphans`.

## Steps

1. Set `UpgradeOp = UpgradeOpUpgradeWarnOrphans`.
2. Set leaf-specific request fields.

## Context

- See ASSERT.md for expected outcomes.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.UpgradeOp = UpgradeOpUpgradeWarnOrphans
	req.OrphanID = "sess-orph01"
	return nil
}
```
