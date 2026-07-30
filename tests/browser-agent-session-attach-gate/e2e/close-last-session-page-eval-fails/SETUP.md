# Scenario

**Feature**: close last session control tab → subsequent eval must not succeed

```
playwright-debug -> open /go?session=S + user tab; eval succeeds (armed)
                 -> close last control tab
POST /v1/jobs eval -> must NOT succeed (no sticky leftover attach / gate)
stdout: {"assert":"eval_fails_after_last_session_page_closed","ok":true,...}
```

## Preconditions

- `PlaywrightOp` = close-last-session-page-eval-fails.
- Session id `sess-attach-gate-close-last`.

## Steps

1. Set `PlaywrightOp = PlaywrightOpCloseLastSessionPageEvalFails`.
2. Set `SessionID = "sess-attach-gate-close-last"`.

## Context

- Behavioral assert: job success/fail after last control tab closed.
- Do not scrape Chrome debugger infobar DOM.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.PlaywrightOp = PlaywrightOpCloseLastSessionPageEvalFails
	req.SessionID = "sess-attach-gate-close-last"
	return nil
}
```
