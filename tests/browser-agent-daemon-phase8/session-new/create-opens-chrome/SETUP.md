# Scenario

**Feature**: SessionNew opens Chrome via injectable hook (no agent-run)

```
SessionNew(explicit id) -> OpenChromeFn once; AgentRunProbeFn never called
```

## Preconditions

- Explicit `SessionID` from root (`sess-new-8`).
- `SessionNewOp` create-opens-chrome.

## Steps

1. Set `SessionNewOp = SessionNewOpCreateOpensChrome`.

## Context

- Requirement: **No AgentRunFn** — agent probe call count must stay 0.
- OpenChrome session URL must reference `/go` and the session id.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.SessionNewOp = SessionNewOpCreateOpensChrome
	return nil
}```
