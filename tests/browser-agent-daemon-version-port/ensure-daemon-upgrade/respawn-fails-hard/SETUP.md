# Scenario

**Feature**: respawn fails hard

```
EnsureDaemon + CompareVersion -> reuse|warn|kill+respawn
```

## Preconditions

- `UpgradeOp = UpgradeOpRespawnFailsHard`.

## Steps

1. Set `UpgradeOp = UpgradeOpRespawnFailsHard`.

## Context

- See ASSERT.md for expected outcomes.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.UpgradeOp = UpgradeOpRespawnFailsHard
	return nil
}
```
