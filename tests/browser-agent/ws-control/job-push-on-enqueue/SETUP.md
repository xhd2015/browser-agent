# Scenario

**Feature**: after hello, enqueue eval → WS receives type=job (D2)

```
Fake Extension hello then Loop
Test Client -> POST /v1/jobs type=eval (background)
Fake Extension <- envelope type=job, payload.type=eval
```

## Preconditions

- Extension connected before POST.
- WSAction job-push.

## Steps

1. Set WSAction to job-push.

## Context

- Assert focuses on WS push; HTTP completion is optional here.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.WSAction = WSActionJobPush
	return nil
}
```
