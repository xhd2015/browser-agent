# Scenario

**Feature**: stale daemon meta with dead pid

```
Write server.json (dead pid) -> QueryDaemonStatus -> Running=false
```

## Preconditions

- `QueryStatusOp` stale-pid.
- Fixture meta written in Run harness (not by QueryDaemonStatus).

## Steps

1. Set `QueryStatusOp = QueryStatusOpStalePID`.
2. Default `StalePID = 999999999` when unset.

## Context

- Stale meta must remain on disk (read-only query).

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.QueryStatusOp = QueryStatusOpStalePID
	return nil
}
```