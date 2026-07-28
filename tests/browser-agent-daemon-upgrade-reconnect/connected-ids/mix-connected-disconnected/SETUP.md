# Scenario

**Feature**: only connected sessions enter waitList; disconnected are skipped

```
ConnectedIDsFromSnapshots([A conn, B disc, C conn]) -> [A, C]
```

## Preconditions

- ConnectedIDsCase = mix-connected-disconnected.

## Steps

1. Seed mix of connected and disconnected snapshots.

## Context

- Disconnected orphans are not in waitList (Phase 1 restore handles dirs separately).

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.ConnectedIDsCase = ConnectedIDsMixConnectedDisconnected
	req.Snapshots = []ConnSnap{
		{SessionID: "sess-a", Connected: true},
		{SessionID: "sess-b", Connected: false},
		{SessionID: "sess-c", Connected: true},
	}
	return nil
}
```
