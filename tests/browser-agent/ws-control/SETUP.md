# Scenario

**Feature**: WebSocket control channel for extension hello / job / result / disconnect

```
Test Client -> browseragent.Run -> Control Server
Fake Extension -> GET /v1/ws
  envelope {v:1, type:hello|job|result|…, id, payload}
Disconnect policy v1: fail inflight jobs (no requeue)
```

## Preconditions

- Mode is `ws-control`.
- Default hello version `1.0.0` and feature `browser-agent`.
- No real Chrome.

## Steps

1. Set `Mode = ModeWSControl`.
2. Default HelloVersion / HelloFeatures.
3. Children set WSAction.

## Context

- Requirement D1–D4. HTTP-only job paths remain under `http-jobs/`.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Mode = ModeWSControl
	if req.HelloVersion == "" {
		req.HelloVersion = "1.0.0"
	}
	if req.HelloFeatures == nil {
		req.HelloFeatures = []string{"browser-agent"}
	}
	return nil
}
```
