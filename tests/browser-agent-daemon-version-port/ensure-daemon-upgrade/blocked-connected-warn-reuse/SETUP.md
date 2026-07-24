# Scenario

**Feature**: blocked connected warn reuse

```
EnsureDaemon + CompareVersion -> reuse|warn|kill+respawn
```

## Preconditions

- `UpgradeOp = UpgradeOpBlockedConnected`.

## Steps

1. Set `UpgradeOp = UpgradeOpBlockedConnected`.

## Context

- See ASSERT.md for expected outcomes.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.UpgradeOp = UpgradeOpBlockedConnected
	return nil
}
```
