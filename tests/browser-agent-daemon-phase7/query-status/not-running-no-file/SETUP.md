# Scenario

**Feature**: absent daemon meta

```
empty BaseDir (no server.json) -> QueryDaemonStatus -> Running=false
```

## Preconditions

- `QueryStatusOp` not-running-no-file.
- No `server.json` written.

## Steps

1. Set `QueryStatusOp = QueryStatusOpNotRunningNoFile`.

## Context

- Read-only; must not create meta file.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.QueryStatusOp = QueryStatusOpNotRunningNoFile
	return nil
}
```