# Scenario

**Feature**: POST `/v1/jobs` unknown session_id → 404

```
POST /v1/jobs session_id=does-not-exist -> 404
```

## Preconditions

- Force unknown session on job POST.

## Steps

1. Set `ForceUnknownSession = true`.
2. Set `UnknownSessionID = "does-not-exist"`.
3. Set `JobHTTPTimeoutMS = 500`.

## Context

- Same wire contract as browser-agent C3, now on registry server.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.ForceUnknownSession = true
	req.UnknownSessionID = "does-not-exist"
	req.JobHTTPTimeoutMS = 500
	return nil
}
```