# Scenario

**Feature**: very large unlikely pid is not alive

```
IsProcessAlive(999999999) -> false
```

## Preconditions

- PID is 999999999 (unlikely to exist).

## Steps

1. Set PID to 999999999.

## Context

- Dead/unassigned pid should return false without error.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.PID = 999999999
	return nil
}
```