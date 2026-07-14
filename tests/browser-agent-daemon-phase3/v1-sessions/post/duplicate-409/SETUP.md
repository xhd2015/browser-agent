# Scenario

**Feature**: POST `/v1/sessions` duplicate session_id → 409

```
POST sess-dup -> 201
POST sess-dup -> 409
```

## Preconditions

- Same session_id posted twice.

## Steps

1. Set `PostSessionID = "sess-dup"`.
2. Enable `PostDuplicateSecond = true`.

## Context

- First response should be 201; second 409 (asserted in leaf).

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.PostSessionID = "sess-dup"
	req.PostDuplicateSecond = true
	return nil
}
```