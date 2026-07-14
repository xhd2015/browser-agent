# Scenario

**Feature**: Run single session; GET snapshot with explicit session query

```
Run(SessionID=sess-p5-compat) -> healthy
GET /v1/session?session=sess-p5-compat -> 200 connected false
```

## Preconditions

- NoOpenChrome, NoAgentRun (via root Setup).

## Steps

1. Set `SessionID = "sess-p5-compat"`.

## Context

- Mirrors phase 3 regression bridge with phase 5 tree session id.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.SessionID = "sess-p5-compat"
	return nil
}
```