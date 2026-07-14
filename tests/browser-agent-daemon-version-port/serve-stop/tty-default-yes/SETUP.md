# Scenario

**Feature**: tty default yes

```
serve --stop + connected -> TTY [Y/n] or non-TTY warn+stop
```

## Preconditions

- `ServeStopOp = ServeStopOpTTYDefaultYes`.

## Steps

1. Set `ServeStopOp = ServeStopOpTTYDefaultYes`.
2. Set leaf-specific request fields.

## Context

- See ASSERT.md for expected outcomes.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.ServeStopOp = ServeStopOpTTYDefaultYes
	req.IsTTY = true
		req.Stdin = "\n"
	return nil
}
```
