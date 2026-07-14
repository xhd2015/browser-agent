# Scenario

**Feature**: GET /preview unknown session → 404 (req #6)

```
Test Client -> GET /preview?session=does-not-exist
Control Server -> HTTP 404 (HTML or JSON not-found)
```

## Preconditions

- Live server is up; probe uses a non-live session id.
- No entry staging required.

## Steps

1. Set `ForceUnknownSession = true`.
2. Set `SessionIDForProbe = "does-not-exist"`.
3. Do not stage posts.

## Context

- Sibling of `session-known/` under preview.
- Error body format flexible (HTML page or JSON).

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.ForceUnknownSession = true
	req.SessionIDForProbe = "does-not-exist"
	req.DoStagePost = false
	req.DoClearAfterStage = false
	return nil
}
```
