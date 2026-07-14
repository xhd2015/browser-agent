# Scenario

**Feature**: POST `/v1/sessions` — create session via registry

```
Test Client -> POST /v1/sessions {session_id}
Registry Control Server -> registry.Create -> 201 | 409 | 400
```

## Preconditions

- Mode `ModeV1SessionsPost`.

## Steps

1. Set `Mode = ModeV1SessionsPost`.
2. Leaves choose success, duplicate, or invalid id.

## Context

- Body `session_id` is explicit in these leaves (deterministic ids).

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Mode = ModeV1SessionsPost
	return nil
}
```