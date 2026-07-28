# Scenario

**Feature**: poll leases a job enqueued after HTTP hello

```
Create + hello sess-ext-poll-job
POST /v1/jobs type=eval (background; no WS)
POST /v1/ext/poll {session_id, wait_ms:2000}
  -> 200 jobs non-empty; first job type=eval; non-empty id
```

## Preconditions

- PollCase = returns-queued-job.
- EnqueueJob = true.
- PollWaitMS = 2000 (allow enqueue race).

## Steps

1. Hello (Run), start background job, poll with wait.

## Context

- Without WS, pushJob skips; job stays Queued until poll leases it.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.PollCase = PollReturnsQueuedJob
	req.SessionID = "sess-ext-poll-job"
	req.EnqueueJob = true
	req.JobHTTPType = "eval"
	req.JobHTTPParams = map[string]any{"expression": "1+1"}
	req.PollWaitMS = 2000
	return nil
}
```
