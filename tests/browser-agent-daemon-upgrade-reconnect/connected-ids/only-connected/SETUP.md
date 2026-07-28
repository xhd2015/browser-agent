# Scenario

**Feature**: all-connected snapshots return every session id in input order

```
ConnectedIDsFromSnapshots([A conn, B conn]) -> [A, B]
```

## Preconditions

- ConnectedIDsCase = only-connected.
- Two connected snapshots.

## Steps

1. Set Snapshots to two connected sessions.

## Context

- Multi-session waitList order preserves input.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.ConnectedIDsCase = ConnectedIDsOnlyConnected
	req.Snapshots = []ConnSnap{
		{SessionID: "sess-aaa", Connected: true},
		{SessionID: "sess-bbb", Connected: true},
	}
	return nil
}
```
