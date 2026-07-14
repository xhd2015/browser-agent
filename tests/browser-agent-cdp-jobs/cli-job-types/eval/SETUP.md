# Scenario

**Feature**: eval posts job type eval with expression (B1)

```
Serve + fake WS
HandleCLI session eval --session-id X --addr <base> '1+1'
  -> observed type=eval
  -> params include expression 1+1 (or expr)
  -> CLI nil error; stdout trailing \n
```

## Preconditions

- JobCLI = eval.
- EvalExpr = 1+1.

## Steps

1. Set JobCLIEval.
2. Set EvalExpr to 1+1.

## Context

- Requirement B1. No real CDP; result from harness fake extension.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.JobCLI = JobCLIEval
	req.EvalExpr = "1+1"
	req.CLIArgs = nil
	return nil
}
```
