# Scenario

**Feature**: Session page DOM in real browser

```
RunDaemon -> POST /v1/sessions
playwright-debug -> goto /go?session=id -> assert session-page markers
```

## Preconditions

- One session per leaf; asserts HTML/DOM not JSON poll alone.

## Steps

1. Leaf sets `PlaywrightOp` and `SessionID`.

## Context

- Sibling `single-session/` focuses on extension.connected poll.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	return nil
}
```