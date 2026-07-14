# Scenario

**Feature**: tty n aborts

```
serve --stop + connected -> TTY [Y/n] or non-TTY warn+stop
```

## Preconditions

- `ServeStopOp = ServeStopOpTTYNAborts`.

## Steps

1. Set `ServeStopOp = ServeStopOpTTYNAborts`.
2. Set leaf-specific request fields.

## Context

- See ASSERT.md for expected outcomes.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.ServeStopOp = ServeStopOpTTYNAborts
	req.IsTTY = true
		req.Stdin = "n\n"
	return nil
}
```
