# Scenario

**Feature**: HandleCLI session info/eval against live in-process control server

```
Test Client -> browseragent.Run(NoOpenChrome, NoAgentRun) @ free port
Fake Extension -> WS hello + auto result (eval)
Operator -> HandleCLI(["session", "eval"|"info", --session-id, --addr, …])
  -> stdout result/snapshot + trailing \n
  -> nil error on success
```

## Preconditions

- Mode is CLI side-command.
- Server uses temp BaseDir + known SessionID + free Addr.
- No real Chrome / agent-run.
- Fake WS used for eval completion (and optionally for info connected field).
- Nested `session` parent only (no flat aliases).

## Steps

1. Set `Mode = ModeCLISidecmd`.
2. Enable FakeExtension default true for eval leaves.
3. Leave Sidecmd to leaf Setup.
4. Increase MaxDispatchWait for job wait.

## Context

- Harness constructs nested `session <cmd>` + --addr and --session-id when CLIArgs empty.
- --addr may be full base URL (`http://127.0.0.1:port`) or host:port — implementer accepts either.

```go
import (
	"testing"
	"time"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Mode = ModeCLISidecmd
	req.NoOpenChrome = true
	req.NoAgentRun = true
	if req.MaxDispatchWait == 0 {
		req.MaxDispatchWait = 8 * time.Second
	}
	if req.CLIEnv == nil {
		req.CLIEnv = map[string]string{}
	}
	return nil
}
```
