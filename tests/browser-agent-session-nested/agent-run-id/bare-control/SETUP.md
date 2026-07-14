# Scenario

**Feature**: bare control id gets agent-run prefix (A1)

```
AgentRunSessionID("demo") -> "browser-agent-sess-demo"
```

## Preconditions

- Control id has no prefix.

## Steps

1. Set ControlSessionID to `demo`.

## Context

- Requirement A1.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.ControlSessionID = "demo"
	return nil
}
```
