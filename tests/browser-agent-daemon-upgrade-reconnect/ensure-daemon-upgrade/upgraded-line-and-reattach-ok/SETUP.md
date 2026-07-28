# Scenario

**Feature**: stderr reports upgraded daemon vOLD → vNEW and reattached session list

```
after kill+spawn + full reattach
  stderr: upgraded daemon v0.1.0 → v0.2.0
  stderr: reattach... sess-r1
```

## Preconditions

- Two connected sessions that reattach post-spawn.

## Steps

1. Set ConnectedSessionIDs = ReattachSessionIDs for full reattach success.

## Context

- Operator-visible confirmation that upgrade and extension reconnect completed.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.EnsureUpgradeCase = EnsureUpgradeLineAndReattachOK
	req.ConnectedSessionIDs = []string{"sess-r1", "sess-r2"}
	req.ReattachSessionIDs = []string{"sess-r1", "sess-r2"}
	req.ClientVersion = "0.2.0"
	req.DaemonVersion = "0.1.0"
	return nil
}
```
