# Scenario

**Feature**: Single session — extension connects via real browser

```
RunDaemon -> POST /v1/sessions (one id, no Chrome)
playwright-debug --extension -> tab /go?session=id
Extension SW -> daemon WS -> GET /v1/session extension.connected true
```

## Preconditions

- One session id per leaf.
- Playwright polls `/v1/session` until `extension.connected` (15s script timeout).

## Steps

1. Leaf sets `PlaywrightOp` and explicit `SessionID`.

## Context

- Sibling `session-page/` asserts DOM; this branch asserts extension JSON poll.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	return nil
}
```