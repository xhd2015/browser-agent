# Scenario

**Feature**: GET `/go?session=<known>` includes session warning banner

```
registry pre-create sess-go
GET /go?session=sess-go -> HTML with data-browser-agent-session-warning + session id
```

## Preconditions

- Known session pre-created.

## Steps

1. Set `SessionID = "sess-go"`.
2. Set `PreCreateSessionIDs = []string{"sess-go"}`.

## Context

- Banner warns about keeping the correct session page open (multi-session UX).

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.SessionID = "sess-go"
	req.PreCreateSessionIDs = []string{"sess-go"}
	return nil
}
```