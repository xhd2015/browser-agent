# Scenario

**Feature**: session control tab open → eval on user tab succeeds (attach allowed)

```
playwright-debug -> Tab1 goto /go?session=S (extension connects)
                 -> Tab2 user URL active
POST /v1/jobs eval -> ok; value.url has LOOP_MARKER
stdout: {"assert":"session_open_eval","ok":true,...}
```

## Preconditions

- `PlaywrightOp` = session-open-eval-succeeds.
- Session id `sess-attach-gate-open`.

## Steps

1. Set `PlaywrightOp = PlaywrightOpSessionOpenEvalSucceeds`.
2. Set `SessionID = "sess-attach-gate-open"`.

## Context

- Baseline armed path: gate must allow attach while control tab is open.
- Script: `testdata/session-open-eval.js`.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.PlaywrightOp = PlaywrightOpSessionOpenEvalSucceeds
	req.SessionID = "sess-attach-gate-open"
	return nil
}
```
