# Scenario

**Feature**: warn older client

```
EnsureDaemon + CompareVersion -> reuse|warn|kill+respawn
```

## Preconditions

- `UpgradeOp = UpgradeOpWarnOlderClient`.

## Steps

1. Set `UpgradeOp = UpgradeOpWarnOlderClient`.

## Context

- See ASSERT.md for expected outcomes.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.UpgradeOp = UpgradeOpWarnOlderClient
	return nil
}
```
