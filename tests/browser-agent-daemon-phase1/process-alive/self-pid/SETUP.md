# Scenario

**Feature**: current process pid is alive

```
IsProcessAlive(os.Getpid()) -> true
```

## Preconditions

- PID is current process (0 means default in Run).

## Steps

1. Leave PID at 0 so Run uses os.Getpid().

## Context

- Self-pid must always be alive during test.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.PID = 0
	return nil
}
```