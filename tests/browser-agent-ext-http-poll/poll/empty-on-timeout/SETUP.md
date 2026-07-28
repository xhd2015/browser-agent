# Scenario

**Feature**: poll with short wait and no work returns empty arrays

```
Create + hello sess-ext-poll-empty
POST /v1/ext/poll {session_id, wait_ms:100}
  -> 200 {jobs:[], events:[]}
```

## Preconditions

- PollCase = empty-on-timeout.
- EnqueueJob = false.
- PollWaitMS = 100.

## Steps

1. Hello only; poll with short wait.
2. Expect empty jobs and events after timeout.

## Context

- Long-poll must not hang forever; empty is success.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.PollCase = PollEmptyOnTimeout
	req.SessionID = "sess-ext-poll-empty"
	req.EnqueueJob = false
	req.PollWaitMS = 100
	return nil
}
```
