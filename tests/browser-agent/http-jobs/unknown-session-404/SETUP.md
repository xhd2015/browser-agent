# Scenario

**Feature**: POST /v1/jobs with unknown session id → 404 (C3)

```
Live session = ba-…
Test Client -> POST /v1/jobs session_id=does-not-exist
  -> HTTP 404
```

## Preconditions

- ForceUnknownSession true.
- SessionIDForProbe `does-not-exist`.

## Steps

1. Force unknown session id on the job POST.
2. No fake extension required.

## Context

- Distinct from resolve-session pure errors (wire-level 404).

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.ForceUnknownSession = true
	req.SessionIDForProbe = "does-not-exist"
	req.FakeExtension = false
	req.JobHTTPTimeoutMS = 500
	return nil
}
```
