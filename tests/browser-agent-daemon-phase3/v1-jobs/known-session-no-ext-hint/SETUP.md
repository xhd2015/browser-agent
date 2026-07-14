# Scenario

**Feature**: POST job to known session without extension → ok:false + disconnected hint

```
registry pre-create sess-job
POST /v1/jobs session_id=sess-job (no WS) -> ok:false + data.hint + data.session_url
```

## Preconditions

- Known session; no fake extension.
- Short server timeout for fast failure.

## Steps

1. Set `SessionID = "sess-job"`.
2. Set `PreCreateSessionIDs = []string{"sess-job"}`.
3. Set `JobHTTPTimeoutMS = 300`.

## Context

- Accept fast-fail "not connected" or timeout if `data.hint` still mentions `/go?session=`.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.SessionID = "sess-job"
	req.PreCreateSessionIDs = []string{"sess-job"}
	req.JobHTTPTimeoutMS = 300
	return nil
}
```