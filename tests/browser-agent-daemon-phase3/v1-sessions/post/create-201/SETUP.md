# Scenario

**Feature**: POST `/v1/sessions` with valid session_id → 201

```
POST {session_id: "sess-create-201"} -> 201 + create JSON
```

## Preconditions

- Fresh registry server; no pre-created sessions.

## Steps

1. Set `PostSessionID = "sess-create-201"`.

## Context

- Response should include `session_id` and `session_url` (or equivalent create result fields).

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.PostSessionID = "sess-create-201"
	req.SessionID = "sess-create-201"
	return nil
}
```