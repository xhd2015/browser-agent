# Scenario

**Feature**: reuse equal version

```
EnsureDaemon + CompareVersion -> reuse|warn|kill+respawn
```

## Preconditions

- `UpgradeOp = UpgradeOpReuseEqual`.

## Steps

1. Set `UpgradeOp = UpgradeOpReuseEqual`.
2. Set leaf-specific request fields.

## Context

- See ASSERT.md for expected outcomes.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.UpgradeOp = UpgradeOpReuseEqual
	req.ClientVersion = "0.2.0"
		req.DaemonVersion = "0.2.0"
	return nil
}
```
