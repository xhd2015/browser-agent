# Scenario

**Feature**: HTTP POST /v1/jobs enqueue-and-wait

```
Test Client -> browseragent.Run (NoOpenChrome, NoAgentRun) -> Control Server
Fake Extension? -> WS /v1/ws
Test Client -> POST /v1/jobs {session_id, type, params, timeout_ms}
Control Server holds until JobResult or timeout | 404 unknown session
```

## Preconditions

- Mode is `http-jobs`.
- Live session id from root Setup unless leaf forces unknown.
- Fake extension only when child enables it.

## Steps

1. Set `Mode = ModeHTTPJobs`.
2. Default job type `eval`.
3. Children choose known vs unknown and extension behavior.

## Context

- Requirement C1–C3. Session id resolution covered under `resolve-session/`.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Mode = ModeHTTPJobs
	if req.JobHTTPType == "" {
		req.JobHTTPType = "eval"
	}
	if req.JobHTTPParams == nil {
		req.JobHTTPParams = map[string]any{"code": "1+1"}
	}
	req.HelloVersion = "1.0.0"
	req.HelloFeatures = []string{"browser-agent"}
	return nil
}
```
