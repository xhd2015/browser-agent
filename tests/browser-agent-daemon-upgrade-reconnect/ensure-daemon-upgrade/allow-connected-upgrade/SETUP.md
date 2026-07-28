# Scenario

**Feature**: client > daemon with extension-connected sessions still kills and respawns

```
connected waitList non-empty + client>daemon
  -> KillFn+SpawnFn called; no "cannot upgrade" reuse of old daemon
```

## Preconditions

- EnsureUpgradeCase = allow-connected-upgrade.
- ConnectedSessionIDs non-empty; client 0.2.0 > daemon 0.1.0.

## Steps

1. Set connected session ids that form the waitList.

## Context

- Removes Phase-0 Q1 block that reused old daemon when connected ≥1.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.EnsureUpgradeCase = EnsureUpgradeAllowConnected
	req.ConnectedSessionIDs = []string{"sess-live01"}
	req.ReattachSessionIDs = []string{"sess-live01"}
	req.ClientVersion = "0.2.0"
	req.DaemonVersion = "0.1.0"
	return nil
}
```
