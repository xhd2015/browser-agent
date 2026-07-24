# Scenario

**Feature**: kill fails hard

```
EnsureDaemon + CompareVersion -> reuse|warn|kill+respawn
```

## Preconditions

- `UpgradeOp = UpgradeOpKillFailsHard`.

## Steps

1. Set `UpgradeOp = UpgradeOpKillFailsHard`.

## Context

- See ASSERT.md for expected outcomes.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.UpgradeOp = UpgradeOpKillFailsHard
	return nil
}
```
