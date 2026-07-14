# Scenario

**Feature**: POST /v1/shutdown HTTP endpoint

```
RunDaemon -> POST /v1/shutdown -> 202 Accepted
RunDaemon -> POST /v1/shutdown -> drain -> exit -> server.json gone
```

## Preconditions

- Mode `ModeV1Shutdown`.
- Leaf sets `V1ShutdownOp`.

## Steps

1. Set `Mode = ModeV1Shutdown`.

## Context

- Uses in-process `RunDaemon`; shutdown triggered by HTTP self-request.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Mode = ModeV1Shutdown
	return nil
}
```