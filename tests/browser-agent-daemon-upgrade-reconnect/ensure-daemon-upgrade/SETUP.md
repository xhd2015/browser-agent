# Scenario

**Feature**: EnsureDaemon client>daemon orchestration — prepare, allow connected, wait reattach, stderr

```
EnsureDaemon(hooks) client>daemon
  -> prepare-upgrade best-effort
  -> sleep delay+slack
  -> kill+spawn (even if connected)
  -> wait reattach ≤30s
  -> stderr upgraded / reattached / still waiting
```

## Preconditions

- Mode = `ensure-daemon-upgrade`.
- Leaf sets `EnsureUpgradeCase` and scenario-specific fields.
- Defaults: client `0.2.0`, daemon `0.1.0`, prepare delay 1000ms, slack 200ms.

## Steps

1. Set `Mode = ModeEnsureDaemonUpgrade`.
2. Apply version and prepare delay defaults when leaf omits them.

## Context

- Uses httptest `/v1/health` + hooks (PrepareUpgradeFn, FetchSessionsFn, SleepFn, KillFn, SpawnFn).
- No real process kill/spawn.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.Mode = ModeEnsureDaemonUpgrade
	if req.ClientVersion == "" {
		req.ClientVersion = "0.2.0"
	}
	if req.DaemonVersion == "" {
		req.DaemonVersion = "0.1.0"
	}
	if req.PrepareDelayMS <= 0 {
		req.PrepareDelayMS = 1000
	}
	// Product default slack ~200ms; leaves may override (including 0).
	if req.PrepareSlackMS == 0 && req.EnsureUpgradeCase == "" {
		// Leaf Setup runs after this grouping Setup and may set case + slack.
		// Apply slack default only when still zero after leaf — handled in Run when
		// leaf left both unset. Set grouping default here for prepare-order leaves.
		req.PrepareSlackMS = 200
	}
	return nil
}
```
