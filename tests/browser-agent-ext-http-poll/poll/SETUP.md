# Scenario

**Feature**: `POST /v1/ext/poll` long-polls jobs and events for a session

```
POST /v1/ext/hello (known session)
POST /v1/jobs (bg) -> job Queued without WS
POST /v1/ext/poll {session_id, wait_ms}
  -> 200 {jobs:[...], events:[...]}
empty wait timeout -> jobs=[] events=[]
prepare_reconnect broadcast -> events include type
unknown session -> 404
```

## Preconditions

- Mode = poll.
- Leaves configure PollCase, wait_ms, enqueue/broadcast flags.

## Steps

1. Set Mode to poll.
2. Default short PollWaitMS for tests when unset at leaf.

## Context

- wait_ms normalized server-side; leaves send explicit short values.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.Mode = ModePoll
	if req.PollWaitMS == 0 && !req.PollOmitWaitMS {
		// Leaves may override; grouping default keeps tests fast when they forget.
		req.PollWaitMS = 500
	}
	return nil
}
```
