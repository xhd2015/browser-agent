# Scenario

**Feature**: empty or nil snapshots yield no connected ids

```
ConnectedIDsFromSnapshots(nil|[]) -> []
```

## Preconditions

- ConnectedIDsCase = empty.
- Snapshots nil.

## Steps

1. Set ConnectedIDsCase and leave Snapshots nil.

## Context

- waitList for upgrade with zero sessions.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.ConnectedIDsCase = ConnectedIDsEmpty
	req.Snapshots = nil
	return nil
}
```
