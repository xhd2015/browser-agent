# Scenario

**Feature**: Two sessions; `/v1/ws?session=sess-p4-a` upgrades

```
registry {sess-p4-a, sess-p4-b}
GET /v1/ws?session=sess-p4-a -> WebSocket upgrade ok
```

## Preconditions

- `WSSessionOp = known-upgrade`.
- Dial targets session A.

## Steps

1. Set `WSSessionOp = WSSessionOpKnownUpgrade`.
2. Set `DialSessionID = req.SessionIDA` (sess-p4-a).

## Context

- Successful upgrade means gorilla Dial returns nil error.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.WSSessionOp = WSSessionOpKnownUpgrade
	req.DialSessionID = req.SessionIDA
	return nil
}
```