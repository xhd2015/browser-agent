# Scenario

**Feature**: poll loop handles prepare_reconnect events from /v1/ext/poll

```
POST /v1/ext/poll
  -> events: [{ type: "prepare_reconnect", delay_ms, … }]
  -> schedule close + aggressive reconnect (shared handler OK)
```

## Preconditions

- BackgroundSourceTarget = prepare-reconnect-on-poll.

## Steps

1. Set `BackgroundSourceTarget = BgSrcPrepareReconnectOnPoll`.

## Context

- `prepare_reconnect` marker required; poll path and/or events drain language required.
- Full delay_ms/close matrix is covered by `tests/browser-agent-prepare-reconnect`;
  this leaf only ties prepare_reconnect to the HTTP poll transport.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.BackgroundSourceTarget = BgSrcPrepareReconnectOnPoll
	return nil
}
```
