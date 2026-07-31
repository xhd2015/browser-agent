# Scenario

**Feature**: regression — empty `sessions.keys()` boot loop is insufficient

```
// OBSOLETE (does not self-heal after cold SW restart):
chrome.runtime.onInstalled.addListener(() => {
  for (const sessionId of sessions.keys()) {
    connectSession(sessionId, "onInstalled");
  }
});

// REQUIRED: tabs.query rediscover of /go?session= + bind + connect (+ re-attach)
```

## Preconditions

- ExtSourceTarget = empty-sessions-keys-insufficient.

## Steps

1. Set `ExtSourceTarget = ExtSrcEmptySessionsKeysInsufficient`.

## Context

- After cold SW restart `sessions` is empty; keys-only loop is a silent no-op.
- This leaf is the explicit **regression** contract: heal must be stronger than
  the old onInstalled/onStartup pattern.
- Satisfied when boot rediscover + reconnect (and thus not keys-only) are present.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.ExtSourceTarget = ExtSrcEmptySessionsKeysInsufficient
	return nil
}
```
