# Scenario

**Feature**: prepare-upgrade failure is best-effort and does not abort the upgrade

```
PrepareUpgradeFn -> error
  -> still KillFn+SpawnFn; EnsureDaemon ok
```

## Preconditions

- PrepareUpgradeErr set to a non-empty error string (simulates old daemon / network fail).

## Steps

1. Inject prepare failure; keep connected waitList optional.

## Context

- Old daemons lack POST /v1/admin/prepare-upgrade; upgrade must continue.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.EnsureUpgradeCase = EnsureUpgradePrepareBestEffort
	req.ConnectedSessionIDs = []string{"sess-be01"}
	req.ReattachSessionIDs = []string{"sess-be01"}
	req.PrepareUpgradeErr = "prepare-upgrade not supported"
	return nil
}
```
