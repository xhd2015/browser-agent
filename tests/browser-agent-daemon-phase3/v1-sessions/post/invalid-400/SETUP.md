# Scenario

**Feature**: POST `/v1/sessions` invalid session_id → 400

```
POST {session_id: "bad/id"} -> 400
```

## Preconditions

- Invalid id per `ValidateSessionID` (contains slash).

## Steps

1. Set `PostInvalidID = "bad/id"`.

## Context

- Must not be 409 (not a duplicate case).

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.PostInvalidID = "bad/id"
	return nil
}
```