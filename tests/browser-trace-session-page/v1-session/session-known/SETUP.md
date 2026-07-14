# Scenario

**Feature**: GET `/v1/session` for the live (known) session id

```
# Probe uses the real session id created by browsertrace.Run
Control Server holds Session id=SessionSuffix
Test Client -> GET /v1/session?session=<SessionSuffix>
Control Server -> 200 JSON snapshot of that session
```

## Preconditions

- `ForceUnknownSession` is false; probe session id is the live `SessionSuffix`.
- Extension staging (hello / recording) is decided by descendants.

## Steps

1. Ensure `ForceUnknownSession = false`.
2. Clear any prior `SessionIDForProbe` override so harness uses the real id.

## Context

- HTTP 200 expected on this branch once the endpoint exists.
- JSON shape is shared; children assert phase/extension/recording fields.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.ForceUnknownSession = false
	req.SessionIDForProbe = ""
	return nil
}
```
