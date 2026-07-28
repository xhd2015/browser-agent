# Scenario

**Feature**: hello connects; poll returns job after enqueue; result completes

```
Create sess-ext-flow-hpr
POST /v1/ext/hello -> phase extension_connected + GET connected
POST /v1/jobs type=eval (bg)
POST /v1/ext/poll wait_ms=3000 -> jobs[0].type=eval
POST /v1/ext/result ok=true -> 200
POST /v1/jobs waiter -> ok=true
```

## Preconditions

- FlowCase = hello-poll-result.
- Full pipeline flags via defaults in Run.

## Steps

1. Set session id and job type for the round-trip.

## Context

- Requirement smoke path for Phase 1.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.FlowCase = FlowHelloPollResult
	req.SessionID = "sess-ext-flow-hpr"
	req.HelloVersion = "1.0.0"
	req.HelloFeatures = []string{"browser-agent"}
	req.JobHTTPType = "eval"
	req.JobHTTPParams = map[string]any{"expression": "2+2"}
	req.PollWaitMS = 3000
	return nil
}
```
