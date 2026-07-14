# Scenario

**Feature**: session eval posts job type eval via fake WS (C4)

```
serve + fake extension
HandleCLI(["session", "eval", --session-id, --addr, "1+1"])
  -> WS job type=eval
  -> CLI nil error preferred; trailing \n
```

## Preconditions

- Live in-process serve; FakeExtension records first job.
- Nested argv only (`session eval`).

## Steps

1. Set CLIKind session-eval-posts-job.
2. Increase MaxDispatchWait for job wait.
3. Leave CLIArgs empty so Run builds nested eval argv.

## Context

- Requirement C4.

```go
import (
	"testing"
	"time"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.CLIKind = CLIKindSessionEvalPostsJob
	req.CLIArgs = nil
	req.FakeExtension = true
	req.EvalExpr = "1+1"
	req.MaxDispatchWait = 10 * time.Second
	return nil
}
```
