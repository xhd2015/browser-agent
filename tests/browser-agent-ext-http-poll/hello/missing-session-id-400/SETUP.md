# Scenario

**Feature**: hello without session_id returns 400

```
POST /v1/ext/hello {}  // omit session_id
  -> 400
```

## Preconditions

- HelloCase = missing-session-id-400.
- HelloOmitSessionID = true.
- Session may still be created (unused).

## Steps

1. Omit session_id from body.
2. Expect 400.

## Context

- Required field guard.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.HelloCase = HelloMissingSessionID400
	req.SessionID = "sess-ext-hello-missing-id"
	req.HelloOmitSessionID = true
	return nil
}
```
