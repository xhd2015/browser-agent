# Scenario

**Feature**: register still creates session entry and starts HTTP poll loop

```
onMessage type=register
  -> sessions[S] = { … }
  -> startHttpPoll / poll loop / connect that uses HTTP poll
     # not only connectSession -> new WebSocket
```

## Preconditions

- BackgroundSourceTarget = register-starts-poll.

## Steps

1. Set `BackgroundSourceTarget = BgSrcRegisterStartsPoll`.

## Context

- Prefer named starters (`startHttpPoll`, `startPollLoop`, …).
- Accept register + poll path + transport/http language when endpoints present.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.BackgroundSourceTarget = BgSrcRegisterStartsPoll
	return nil
}
```
