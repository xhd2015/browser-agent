# Scenario

**Feature**: running daemon status includes daemon version

```
RunDaemon -> POST /v1/sessions -> QueryDaemonStatus -> DaemonVersion non-empty
```

## Preconditions

- `QueryStatusOp` running-populates-version.

## Steps

1. Set `QueryStatusOp = QueryStatusOpRunningPopulatesVersion`.

## Context

- Version from meta/health/effective chain when daemon is up.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.QueryStatusOp = QueryStatusOpRunningPopulatesVersion
	return nil
}
```