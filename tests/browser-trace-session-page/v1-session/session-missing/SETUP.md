# Scenario

**Feature**: unknown session id on GET `/v1/session` returns not-found

```
# Live session exists, but client queries a different id
Control Server holds Session id=real
Test Client -> GET /v1/session?session=does-not-exist
Control Server -> HTTP 404 JSON (not found)
```

## Preconditions

- A control server session is running with a real `SessionSuffix` id.
- The probe uses a **different** session id that is not registered.
- No hello/status staging is required for this branch (lookup fails first).

## Steps

1. Set `ForceUnknownSession = true`.
2. Use default bogus id `does-not-exist` (or leave `SessionIDForProbe` empty for harness default).
3. Do not post hello or status.

## Context

- Product choice: **HTTP 404** for unknown non-empty session id (not 400).
- Body must indicate not found (message/error field or similar).

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.ForceUnknownSession = true
	req.SessionIDForProbe = "does-not-exist"
	req.DoHello = false
	req.DoStatusRecording = false
	return nil
}
```
