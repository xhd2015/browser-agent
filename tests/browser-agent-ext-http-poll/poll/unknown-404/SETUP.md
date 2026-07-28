# Scenario

**Feature**: poll for unknown session returns 404

```
POST /v1/ext/poll {session_id:"no-such-session-ext-poll", wait_ms:50}
  -> 404
```

## Preconditions

- PollCase = unknown-404.
- OmitCreate = true.
- PollSessionID = no-such-session-ext-poll.

## Steps

1. Do not create target session.
2. POST poll with unknown id.

## Context

- Guard against poll storms on typos.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.PollCase = PollUnknown404
	req.OmitCreate = true
	req.SessionID = "sess-unused-poll-404"
	req.PollSessionID = "no-such-session-ext-poll"
	req.PollWaitMS = 50
	return nil
}
```
