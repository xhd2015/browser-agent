# Scenario

**Feature**: all-disconnected snapshots yield empty waitList

```
ConnectedIDsFromSnapshots([A disc, B disc]) -> []
```

## Preconditions

- ConnectedIDsCase = all-disconnected.

## Steps

1. Seed only disconnected snapshots.

## Context

- Upgrade may still proceed with empty waitList (no reattach wait work).

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.ConnectedIDsCase = ConnectedIDsAllDisconnected
	req.Snapshots = []ConnSnap{
		{SessionID: "sess-x", Connected: false},
		{SessionID: "sess-y", Connected: false},
	}
	return nil
}
```
