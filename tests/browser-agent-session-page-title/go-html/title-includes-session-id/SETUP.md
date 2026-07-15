# Scenario

**Feature**: GET `/go?session=sess-title` HTML title includes session id (T1)

```
registry pre-create sess-title
GET /go?session=sess-title
  -> 200 HTML
  -> <title>sess-title - Browser Agent</title>
  -> not sole static "Browser Agent Session"
```

## Preconditions

- Known session pre-created via registry.

## Steps

1. Set `SessionID = "sess-title"`.
2. Set `PreCreateSessionIDs = []string{"sess-title"}`.

## Context

- Requirement scenario 1 (inject SPA path).
- Exact format: `sessionId + " - Browser Agent"` (spaces around hyphen).

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.SessionID = "sess-title"
	req.PreCreateSessionIDs = []string{"sess-title"}
	return nil
}
```
