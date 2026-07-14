# Scenario

**Feature**: serve custom host port

```
serve --host/--port -> bind 127.0.0.1:N
EnsureDaemon spawn -> default port (no :0)
```

## Preconditions

- `PortOp = PortOpServeCustomHostPort`.

## Steps

1. Set `PortOp = PortOpServeCustomHostPort`.

## Context

- See ASSERT.md for expected outcomes.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.PortOp = PortOpServeCustomHostPort
	return nil
}
```
