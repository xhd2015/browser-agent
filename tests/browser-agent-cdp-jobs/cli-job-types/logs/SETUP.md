# Scenario

**Feature**: logs posts job type logs (B3)

```
Serve + fake WS
HandleCLI session logs --session-id X --addr <base>
  -> observed type=logs
  -> CLI nil error; stdout trailing \n
```

## Preconditions

- JobCLI = logs.

## Steps

1. Set JobCLILogs.

## Context

- Requirement B3. Optional limit/level params not required.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.JobCLI = JobCLILogs
	req.CLIArgs = nil
	return nil
}
```
