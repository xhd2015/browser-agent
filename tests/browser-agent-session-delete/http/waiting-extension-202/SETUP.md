# Scenario

**Feature**: DELETE waiting session returns success

```
POST /v1/sessions (no WS)
DELETE /v1/session?session=ID -> 200 or 204
GET /v1/sessions -> id absent
```

## Preconditions

- No extension connection.

## Steps

1. Set `HTTPOp = waiting-extension-202`.
2. `ConnectExtension = false`.

## Context

- HTTP delete on live daemon control handler.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.HTTPOp = HTTPOpWaiting202
	req.ConnectExtension = false
	return nil
}
```