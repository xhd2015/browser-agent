# Scenario

**Feature**: running daemon status includes sessions

```
RunDaemon -> POST /v1/sessions -> QueryDaemonStatus -> Running=true + sessions
```

## Preconditions

- `QueryStatusOp` running-with-sessions.

## Steps

1. Set `QueryStatusOp = QueryStatusOpRunningWithSessions`.

## Context

- Session id from root Setup (`sess-status-7`).

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.QueryStatusOp = QueryStatusOpRunningWithSessions
	return nil
}
```