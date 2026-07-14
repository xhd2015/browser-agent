# Scenario

**Feature**: DELETE rejected when extension connected

```
POST /v1/sessions -> Fake Extension hello
DELETE /v1/session?session=ID -> 409 extension connected
```

## Preconditions

- Fake extension hello before DELETE.

## Steps

1. Set `HTTPOp = connected-409`.
2. `ConnectExtension = true`.

## Context

- Extension socket stays open during DELETE.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.HTTPOp = HTTPOpConnected409
	req.ConnectExtension = true
	return nil
}
```