# Scenario

**Feature**: SessionNew prints instructions before wait; wait progress/timeout on stderr

```
SessionNew(NoWait=false, short timeout) -> stdout recipes then stderr Waiting… + warning
```

## Preconditions

- Explicit session id `sess-wait-8`.
- No real Chrome extension (wait always times out).

## Steps

1. Set `SessionNewOp = SessionNewOpWaitTimeoutProgress`.
2. Set `SessionID = "sess-wait-8"`.
3. Set `NoWait = false`, `WaitExtensionTimeout = 1500ms`.

## Context

- Operator instructions (session-id, Session URL, Next recipes) must appear on **stdout** even when the extension never connects.
- Wait line and timeout warning go to **stderr** after instructions.

```go
import (
	"testing"
	"time"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.SessionNewOp = SessionNewOpWaitTimeoutProgress
	req.SessionID = "sess-wait-8"
	noWait := false
	req.NoWait = &noWait
	req.WaitExtensionTimeout = 1500 * time.Millisecond
	return nil
}
```
