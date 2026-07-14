# Scenario

**Feature**: background opens per-session WebSocket (P1)

```
Background on register
  -> new WebSocket(ws://127.0.0.1:PORT/v1/ws?session=<id>)
  -> hello on that socket
```

## Preconditions

- ExtSourceTarget = background-per-session-ws.

## Steps

1. Set ExtSourceTarget background-per-session-ws.

## Context

- Single global `/v1/ws` without session query is insufficient for multi-session daemon.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.ExtSourceTarget = ExtSrcBackgroundPerSessionWS
	return nil
}
```