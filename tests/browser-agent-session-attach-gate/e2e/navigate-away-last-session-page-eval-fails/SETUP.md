# Scenario

**Feature**: navigate away last session control tab → subsequent eval must not succeed

```
playwright-debug -> open /go?session=S + user tab; eval succeeds
                 -> navigate control tab away from /go?session=
POST /v1/jobs eval -> must NOT succeed
stdout: {"assert":"eval_fails_after_last_session_page_navigated_away","ok":true,...}
```

## Preconditions

- `PlaywrightOp` = navigate-away-last-session-page-eval-fails.
- Session id `sess-attach-gate-nav-away`.

## Steps

1. Set `PlaywrightOp = PlaywrightOpNavigateAwayLastSessionPageEvalFails`.
2. Set `SessionID = "sess-attach-gate-nav-away"`.

## Context

- Same gate as close for leave policy; navigate-away must recount remaining = 0.
- Behavioral job assert only (no infobar scraping).

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.PlaywrightOp = PlaywrightOpNavigateAwayLastSessionPageEvalFails
	req.SessionID = "sess-attach-gate-nav-away"
	return nil
}
```
