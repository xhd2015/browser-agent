# Scenario

**Feature**: `QueryDaemonStatus` read-only probe

```
QueryDaemonStatus(baseDir) -> DaemonStatus
  running: live pid + GET /v1/sessions
  not running: no server.json
  stale: dead pid in server.json
```

## Preconditions

- Mode `ModeQueryStatus`.
- Leaf sets `QueryStatusOp`.

## Steps

1. Set `Mode = ModeQueryStatus`.

## Context

- Must not mutate `server.json` after query.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Mode = ModeQueryStatus
	return nil
}
```