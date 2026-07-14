# Scenario

**Feature**: Two sessions; `/v1/ws` without `session` → 400

```
registry {sess-p4-a, sess-p4-b}
GET /v1/ws (no query) -> 400 missing session
```

## Preconditions

- `WSSessionOp = missing-400`.
- Omit `session` on WS dial.

## Steps

1. Set `WSSessionOp = WSSessionOpMissing400`.
2. Set `OmitWSSessionParam = true`.

## Context

- Must not upgrade or fall back to an arbitrary session when count ≥ 2.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.WSSessionOp = WSSessionOpMissing400
	req.OmitWSSessionParam = true
	return nil
}
```