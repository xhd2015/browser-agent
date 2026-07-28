# Scenario

**Feature**: `POST /v1/ext/result` completes a polled job like WS result

```
hello + POST /v1/jobs (bg) + poll -> job_id
POST /v1/ext/result {session_id, job_id, ok:true, data?}
  -> 200 {ok:true}
  -> /v1/jobs waiter ok=true
unknown session -> 404
```

## Preconditions

- Mode = result.

## Steps

1. Set Mode to result.
2. Leaves set ResultCase.

## Context

- Completes via same JobQueue.Complete path as handleWSResult.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.Mode = ModeResult
	req.ResultOK = true
	if req.PollWaitMS <= 0 {
		req.PollWaitMS = 2000
	}
	return nil
}
```
