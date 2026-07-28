# Scenario

**Feature**: Single orphan id survives ensureDaemonKillAndRespawn

```
seed sess-orph01 + marker -> kill+respawn([sess-orph01])
  -> dir remains; marker intact
```

## Preconditions

- UpgradeKeepCase = kill-respawn-orphans-remain.
- OrphanIDs = [`sess-orph01`].
- Seed with Marker `keep-me`.

## Steps

1. Set UpgradeKeepCase, OrphanIDs, Seeds.

## Context

- Core Phase 1 upgrade policy: keep dirs for restore on respawn.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.UpgradeKeepCase = UpgradeKeepSingleOrphan
	req.OrphanIDs = []string{"sess-orph01"}
	req.Seeds = []SeedSession{
		{ID: "sess-orph01", Marker: "keep-me"},
	}
	return nil
}
```
