# Scenario

**Feature**: duplicate session id rejected on second SessionNew

```
SessionNew(sess-dup-8) -> ok
SessionNew(sess-dup-8) -> duplicate / 409 error
```

## Preconditions

- Fixed session id `sess-dup-8`.
- `SessionNewOp` duplicate-409.

## Steps

1. Set `SessionNewOp = SessionNewOpDuplicate409`.
2. Set `SessionID = "sess-dup-8"`.

## Context

- First call must succeed; second must fail with duplicate/exists/409 in error text.
- Only one session for id on server.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.SessionNewOp = SessionNewOpDuplicate409
	req.SessionID = "sess-dup-8"
	return nil
}```
