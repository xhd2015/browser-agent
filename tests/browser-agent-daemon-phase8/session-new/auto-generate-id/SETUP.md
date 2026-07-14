# Scenario

**Feature**: auto-generate session id when `--session-id` omitted

```
SessionNew(empty SessionID) -> generated ^sess-[a-z0-9]{6}$ + registered on server
```

## Preconditions

- No explicit session id.
- `SessionNewOp` auto-generate-id.

## Steps

1. Set `SessionNewOp = SessionNewOpAutoGenerateID`.
2. Clear `SessionID = ""`.

## Context

- Generated id must appear in stdout and `GET /v1/sessions`.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.SessionNewOp = SessionNewOpAutoGenerateID
	req.SessionID = ""
	return nil
}```
