# Scenario

**Feature**: HandleCLI dispatch for create-tab without live control server

```
# Dispatch-only: no listen socket, no Chrome
Operator -> HandleCLI(args, empty env, stdout, stderr)
  --help / -h -> help listing create-tab + nil error
  session create-tab without session -> error names both session sources
```

## Preconditions

- Mode is CLI dispatch.
- No server start; MaxDispatchWait bounds accidental hang.
- CLIEnv empty map (no ambient BROWSER_AGENT_SESSION_ID).

## Steps

1. Set `Mode = ModeCLIDispatch`.
2. Clear CLIEnv to empty map when unset.
3. Leave DispatchKind / CLIArgs to leaf Setup.

## Context

- Asserts inspect CLIErr + combined stdout/stderr.
- Transport error from Run is only timeout; expected CLI failures return resp with CLIErr.

```go
import (
	"testing"
	"time"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Mode = ModeCLIDispatch
	if req.CLIEnv == nil {
		req.CLIEnv = map[string]string{}
	}
	if req.MaxDispatchWait == 0 {
		req.MaxDispatchWait = 3 * time.Second
	}
	return nil
}
```
