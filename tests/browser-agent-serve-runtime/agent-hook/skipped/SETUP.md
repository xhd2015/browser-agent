# Scenario

**Feature**: NoAgentRun=true → AgentRunFn not called (C1)

```
Config{NoAgentRun:true, AgentRunFn:record} -> after health: callCount==0
```

## Preconditions

- HookExpect = skipped.

## Steps

1. Set HookExpect skipped.
2. Set NoAgentRun true.

## Context

- Flag must short-circuit before injector and production agent-run.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.HookExpect = HookExpectSkipped
	req.NoAgentRun = true
	return nil
}
```
