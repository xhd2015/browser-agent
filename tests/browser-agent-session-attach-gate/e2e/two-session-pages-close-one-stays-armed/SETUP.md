# Scenario

**Feature**: two session control tabs; close one → stay armed; eval still succeeds

```
playwright-debug -> Tab1 /go?session=S; Tab2 /go?session=S; Tab3 user
                 -> close Tab2 (last registered entry.tabId under master)
POST /v1/jobs eval -> must still succeed (remaining control tab keeps session armed)
stdout: {"assert":"stays_armed_after_one_of_two_closed","ok":true,...}
```

## Preconditions

- `PlaywrightOp` = two-session-pages-close-one-stays-armed.
- Session id `sess-attach-gate-two-pages`.

## Steps

1. Set `PlaywrightOp = PlaywrightOpTwoSessionPagesCloseOneStaysArmed`.
2. Set `SessionID = "sess-attach-gate-two-pages"`.

## Context

- Classic TDD RED on current master: leave only watches `entry.tabId`, so closing
  the last-registered control tab tears down the session even when another
  control tab remains.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.PlaywrightOp = PlaywrightOpTwoSessionPagesCloseOneStaysArmed
	req.SessionID = "sess-attach-gate-two-pages"
	return nil
}
```
