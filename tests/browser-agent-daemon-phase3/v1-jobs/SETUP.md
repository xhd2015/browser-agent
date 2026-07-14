# Scenario

**Feature**: POST `/v1/jobs` — multi-session job RPC

```
Test Client -> POST /v1/jobs {session_id, type, params, timeout_ms}
Registry Control Server -> enqueue/wait | 404 unknown | ok:false + data.hint
```

## Preconditions

- `session_id` required in body for multi-session server.

## Steps

1. Set `Mode = ModeV1Jobs`.
2. Default job type `eval`, short timeout for no-extension leaf.

## Context

- Fake extension not used in phase 3 leaves (no-ext hint path).

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Mode = ModeV1Jobs
	if req.JobHTTPType == "" {
		req.JobHTTPType = "eval"
	}
	if req.JobHTTPParams == nil {
		req.JobHTTPParams = map[string]any{"code": "1+1"}
	}
	return nil
}
```