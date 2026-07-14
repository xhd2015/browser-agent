# Scenario

**Feature**: shutdown endpoint stops daemon host

```
RunDaemon -> POST /v1/shutdown -> RunDaemon exits -> server.json absent
```

## Preconditions

- `V1ShutdownOp` server-stops.
- `server.json` present before POST.

## Steps

1. Set `V1ShutdownOp = V1ShutdownOpServerStops`.

## Context

- Must not use ctx cancel; only HTTP shutdown drives exit.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.V1ShutdownOp = V1ShutdownOpServerStops
	return nil
}
```