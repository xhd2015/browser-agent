# Scenario

**Bug**: `session eval` without `--addr` must resolve addr from `server.json`

```
RunDaemon -> server.json
POST /v1/sessions
HandleCLI session eval --session-id --base-dir '1+1'   # omit --addr
  -> POST /v1/jobs on correct daemon (not 404 unknown session)
```

## Preconditions

- Sidecmd = eval for this subtree.
- No fake WebSocket — job may fail on disconnected extension; 404 unknown session is the bug.

## Steps

1. Set Sidecmd SidecmdEval.
2. Child leaf sets AddrSource from-server-json.

## Context

- Same addr resolution path as `session info`; asserts job routing not session lookup 404.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Sidecmd = SidecmdEval
	req.EvalExpr = "1+1"
	return nil
}
```