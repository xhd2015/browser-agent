# Scenario

**Feature**: already-prefixed control id is idempotent (A2)

```
AgentRunSessionID("browser-agent-sess-demo") -> "browser-agent-sess-demo"
# no double prefix browser-agent-sess-browser-agent-sess-demo
```

## Preconditions

- Control id already starts with AgentRunSessionIDPrefix.

## Steps

1. Set ControlSessionID to `browser-agent-sess-demo`.

## Context

- Requirement A2.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.ControlSessionID = "browser-agent-sess-demo"
	return nil
}
```
