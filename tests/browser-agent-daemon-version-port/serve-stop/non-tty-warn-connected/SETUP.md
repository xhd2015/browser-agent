# Scenario

**Feature**: non tty warn connected

```
serve --stop + connected -> TTY [Y/n] or non-TTY warn+stop
```

## Preconditions

- `ServeStopOp = ServeStopOpNonTTYWarnConnected`.

## Steps

1. Set `ServeStopOp = ServeStopOpNonTTYWarnConnected`.
2. Set leaf-specific request fields.

## Context

- See ASSERT.md for expected outcomes.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.ServeStopOp = ServeStopOpNonTTYWarnConnected
	req.IsTTY = false
	return nil
}
```
