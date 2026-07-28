# Scenario

**Feature**: connected sessions no longer block upgrade (Phase 3)

```
EnsureDaemon + CompareVersion + connected
  -> prepare + kill+respawn (not reuse / cannot-upgrade)
  SessionNew still creates session B
```

## Preconditions

- `UpgradeOp = UpgradeOpBlockedConnected` (historical op name; behavior is allow-upgrade).

## Steps

1. Set `UpgradeOp = UpgradeOpBlockedConnected`.

## Context

- Pre-Phase-3 Q1 reused the old daemon when extension-connected ≥1.
- Phase 3 removes that block; ASSERT expects upgrade + session create.

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
