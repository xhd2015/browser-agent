# Scenario

**Feature**: Run single session; GET snapshot with explicit session query

```
Run(SessionID=sess-run-bridge) -> healthy
GET /v1/session?session=sess-run-bridge -> 200 connected false
```

## Preconditions

- NoOpenChrome, NoAgentRun (via root Setup + Run harness).

## Steps

1. Set `SessionID = "sess-run-bridge"`.

## Context

- Mirrors browser-agent session-ui probes with explicit `?session=` after multi-session refactor.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.SessionID = "sess-run-bridge"
	return nil
}
```