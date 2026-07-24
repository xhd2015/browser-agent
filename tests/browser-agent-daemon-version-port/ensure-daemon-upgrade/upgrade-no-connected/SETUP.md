# Scenario

**Feature**: upgrade no connected

```
EnsureDaemon + CompareVersion -> reuse|warn|kill+respawn
```

## Preconditions

- `UpgradeOp = UpgradeOpUpgradeNoConnected`.

## Steps

1. Set `UpgradeOp = UpgradeOpUpgradeNoConnected`.

## Context

- See ASSERT.md for expected outcomes.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.UpgradeOp = UpgradeOpUpgradeNoConnected
	return nil
}
```
