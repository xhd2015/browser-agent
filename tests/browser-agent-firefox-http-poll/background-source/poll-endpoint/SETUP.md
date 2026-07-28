# Scenario

**Feature**: background long-polls via POST /v1/ext/poll with wait_ms

```
loop:
  POST /v1/ext/poll { session_id, wait_ms }
    -> { jobs: [...], events: [...] }
```

## Preconditions

- BackgroundSourceTarget = poll-endpoint.

## Steps

1. Set `BackgroundSourceTarget = BgSrcPollEndpoint`.

## Context

- Path token `/v1/ext/poll` required; `wait_ms` / `waitMs` required for long-poll.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.BackgroundSourceTarget = BgSrcPollEndpoint
	return nil
}
```
