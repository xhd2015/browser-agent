# Scenario

**Feature**: background maintains sessions map on register (P2)

```
Content Script -> sendMessage({type:"register", session_id, tabId, windowId})
Background on register -> sessions Map<sessionId, {ws, tabId, windowId}>
```

## Preconditions

- ExtSourceTarget = background-session-map.

## Steps

1. Set ExtSourceTarget background-session-map.

## Context

- Background must track tab/window per session for job routing and unregister on close.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.ExtSourceTarget = ExtSrcBackgroundSessionMap
	return nil
}
```