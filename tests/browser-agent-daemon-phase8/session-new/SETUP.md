# Scenario

**Feature**: `SessionNew` package API

```
EnsureDaemon -> POST /v1/sessions -> OpenChromeFn -> pretty stdout
# never agent-run
```

## Preconditions

- Mode `ModeSessionNew`.
- Leaf sets `SessionNewOp`.
- Recording `OpenChromeFn` and `AgentRunProbeFn` injected by harness `Run`.

## Steps

1. Set `Mode = ModeSessionNew`.

## Context

- No pre-started daemon unless duplicate leaf needs same server (SessionNew calls EnsureDaemon).
- Explicit session id default from root: `sess-new-8`.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Mode = ModeSessionNew
	return nil
}```
