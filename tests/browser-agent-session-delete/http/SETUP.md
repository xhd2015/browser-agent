# Scenario

**Feature**: DELETE /v1/session on registry control handler

```
RunDaemon -> POST /v1/sessions
DELETE /v1/session?session=ID -> 200|204 or 409
GET /v1/sessions -> id absent after success
```

## Preconditions

- Mode `http`; live daemon with created session.

## Steps

1. Set `Mode = http`.
2. Leaf Setup sets `HTTPOp` and extension-connect flag.

## Context

- Connected leaf dials fake extension hello before DELETE.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Mode = ModeHTTP
	return nil
}
```