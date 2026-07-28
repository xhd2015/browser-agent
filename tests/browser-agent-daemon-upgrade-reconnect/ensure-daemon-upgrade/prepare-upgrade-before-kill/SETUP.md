# Scenario

**Feature**: prepare-upgrade is invoked with delay_ms=1000 before pre-kill sleep and kill

```
PrepareUpgradeFn(baseURL, 1000) -> SleepFn(delay+slack) -> KillFn
```

## Preconditions

- EnsureUpgradeCase = prepare-upgrade-before-kill.
- PrepareDelayMS=1000, PrepareSlackMS=200.
- Zero or one connected id (order still holds with empty waitList).

## Steps

1. Set prepare delay/slack for sleep-order assertion.
2. Optional single connected id.

## Context

- Product sequence: notify extensions, wait for them to close, then kill.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.EnsureUpgradeCase = EnsureUpgradePrepareBeforeKill
	req.ConnectedSessionIDs = []string{"sess-prep01"}
	req.ReattachSessionIDs = []string{"sess-prep01"}
	req.PrepareDelayMS = 1000
	req.PrepareSlackMS = 200
	return nil
}
```
