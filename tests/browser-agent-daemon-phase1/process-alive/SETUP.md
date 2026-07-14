# Scenario

**Feature**: IsProcessAlive checks whether a process exists

```
IsProcessAlive(pid) -> true | false
```

## Preconditions

- Mode is process-alive.
- Leaf sets PID under test.

## Steps

1. Set Mode to process-alive.

## Context

- Portable check via signal 0 or equivalent.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Mode = ModeProcessAlive
	return nil
}
```