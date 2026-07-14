# Scenario

**Feature**: session eval against live server with fake WS completion (B1)

```
Serve session X + fake extension auto-completes jobs
HandleCLI session eval --session-id X --addr <base> '1+1'
  -> nil error
  -> stdout contains result / ok markers
  -> trailing \n
```

## Preconditions

- Sidecmd = eval; FakeExtension true.
- Expression default 1+1.
- Nested `session eval` only.

## Steps

1. Set SidecmdEval.
2. Enable FakeExtension.
3. Set EvalExpr to 1+1.

## Context

- No real CDP; result payload from harness fake extension.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Sidecmd = SidecmdEval
	req.FakeExtension = true
	req.EvalExpr = "1+1"
	// CLIArgs left empty so Run injects session eval + --session-id and --addr.
	req.CLIArgs = nil
	return nil
}
```
