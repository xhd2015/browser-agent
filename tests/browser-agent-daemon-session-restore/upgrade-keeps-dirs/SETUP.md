# Scenario

**Feature**: Normal upgrade kill+respawn must not remove orphan session dirs

```
seed sessions/{orphan}/ + preserved.txt
ensureDaemonKillAndRespawn(cfg, meta, stderr, orphanIDs)
  -> SessionDirExists true; preserved.txt content keep-me
```

## Preconditions

- Mode = upgrade-keeps-dirs.
- Orphan session dirs seeded with marker file `preserved.txt` = `keep-me`.
- KillFn/SpawnFn no-op (pure wipe-policy test; no real daemon).

## Steps

1. Set `Mode = ModeUpgradeKeepsDirs`.
2. Leaves set UpgradeKeepCase, OrphanIDs, Seeds with Marker.

## Context

- Current code calls `removeSessionDirs` — Phase 1 removes that call for normal upgrade.
- Tests use `TestExported_ensureDaemonKillAndRespawn` when helper stays unexported.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.Mode = ModeUpgradeKeepsDirs
	ensureBaseDir(t, req)
	ensureAddr(t, req)
	return nil
}
```
