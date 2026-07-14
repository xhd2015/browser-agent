# Scenario

**Feature**: SessionNew skips OpenChrome when NoOpenChrome is set

```
SessionNew(NoOpenChrome=true) -> POST /v1/sessions + pretty stdout
# OpenChromeFn never called; never agent-run
```

## Preconditions

- Explicit `SessionID` from root (`sess-new-8`).
- `SessionNewOp` skip-open-chrome.
- Leaf sets `NoOpenChrome = true` on `Request` for harness `Run`.

## Steps

1. Set `SessionNewOp = SessionNewOpSkipOpenChrome`.
2. Set `NoOpenChrome = true`.

## Context

- Session creation and operator stdout must succeed without opening Chrome.
- `OpenChromeFn` and `AgentRunProbeFn` call counts must remain 0.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.SessionNewOp = SessionNewOpSkipOpenChrome
	req.NoOpenChrome = true
	return nil
}
```