# Scenario

**Feature**: heal path reconnects WS for rediscovered sessions

```
heal / onInstalled rediscover
  -> for each open /go?session=S tab:
       getOrCreateSessionEntry(S) / handleRegisterMessage / maybeRegisterGoTab
       bind tabId + windowId
       connectSession(S, "heal"|"onInstalled"|…)
```

## Preconditions

- ExtSourceTarget = reconnect-ws-on-heal.

## Steps

1. Set `ExtSourceTarget = ExtSrcReconnectWSOnHeal`.

## Context

- Current `onInstalled` calls `connectSession` only for `sessions.keys()` — empty
  after cold boot — **RED** until rediscover feeds connect/register.
- Prefer routing through `maybeRegisterGoTab` / `handleRegisterMessage` so bind +
  connect + (P2) eager attach stay consistent; or explicit bind + `connectSession`.
- `connectSession` inside the empty-keys loop alone does **not** satisfy.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.ExtSourceTarget = ExtSrcReconnectWSOnHeal
	return nil
}
```
